package demojify

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// ScanConfig configures how [ScanDir] walks and checks files in a directory
// tree. Use [DefaultScanConfig] for sensible defaults.
type ScanConfig struct {
	// Root is the directory to scan. An empty string is treated as ".".
	Root string

	// SkipDirs lists directory names (or trailing-slash variants) to skip
	// during scanning. A directory is skipped whenever its name matches an
	// entry exactly, regardless of how deep it is in the tree -- so "vendor/"
	// skips both a top-level vendor/ directory and any nested subdir/vendor/.
	// Use forward slashes regardless of operating system (e.g., ".git/",
	// "vendor/", "node_modules/").
	SkipDirs []string

	// ExemptFiles lists base filenames that are exempt from checks.
	// A file whose base name matches any entry in this slice is skipped
	// entirely (e.g., "README.md").
	ExemptFiles []string

	// ExemptSuffixes lists file name suffixes that are exempt from checks.
	// For example, "_test.go" exempts all Go test files.
	ExemptSuffixes []string

	// Extensions filters which file types to scan. Only files whose names
	// end with one of these extensions are checked. An empty or nil slice
	// (the default) means all files are checked regardless of extension.
	// Set this to a specific list (e.g., []string{".go", ".md"}) to
	// restrict scanning to those file types only.
	Extensions []string

	// MaxFileBytes is the maximum file size in bytes that ScanDir will read.
	// Files larger than this value are silently skipped. A value of zero
	// disables the limit and all files are read regardless of size.
	// DefaultScanConfig sets this to 1 MiB (1 << 20).
	MaxFileBytes int64

	// Options configures the sanitization pipeline applied to each file's
	// content. See [Options] and [DefaultOptions].
	Options Options

	// Replacements maps emoji (or other Unicode) sequences to text substitutes.
	// When non-nil, [ScanDir] cleans files with [Replace] instead of [Sanitize],
	// so matched sequences are substituted before any residual emoji are stripped.
	// Longer keys are matched before shorter ones (variation-selector aware).
	// Has no effect on [ScanFile], which accepts only [Options].
	//
	// NOTE: When Replacements is set, [Options.RemoveEmojis] and
	// [Options.AllowedRanges] are ignored because [Replace] always strips
	// residual emoji via [Demojify]. Only [Options.NormalizeWhitespace] is
	// applied after substitution. To preserve specific Unicode ranges during
	// replacement-based scans, add those codepoints to the Replacements map
	// with identity values (key == value).
	Replacements map[string]string

	// CollectMatches controls whether [ScanDir] populates [Finding.Matches] for
	// each file that has emoji. When true, every emoji codepoint occurrence is
	// recorded with its line, column, and surrounding context. Setting this flag
	// carries a small per-file allocation cost; leave it false for bulk audits
	// that only need the cleaned text.
	CollectMatches bool
}

// DefaultScanConfig returns a ScanConfig suitable for auditing a typical Go
// module repository or AI-agent workspace. It skips .git/, vendor/, and
// node_modules/ directories, exempts *_test.go files, and scans all file
// types by default (Extensions is nil). Files larger than 1 MiB are skipped
// to avoid reading large files into memory, and binary files (detected by a
// NUL byte in the first 512 bytes) are always skipped regardless of size.
//
// To restrict scanning to specific extensions, set Extensions explicitly:
//
//	cfg := demojify.DefaultScanConfig()
//	cfg.Extensions = []string{".go", ".md"}
//
// Whitespace normalization is disabled because file formatters (gofmt,
// editors) own trailing-newline conventions.
func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		Root:           ".",
		SkipDirs:       []string{".git/", "vendor/", "node_modules/"},
		ExemptSuffixes: []string{"_test.go"},
		Extensions:     nil,     // scan all file types
		MaxFileBytes:   1 << 20, // 1 MiB
		Options: Options{
			RemoveEmojis: true,
		},
	}
}

// Match describes a single emoji occurrence within a scanned file.
// It is populated in [Finding.Matches] when [ScanConfig.CollectMatches] is true.
type Match struct {
	// Emoji is the matched codepoint sequence (e.g., U+2705 for a check mark).
	Emoji string

	// Replacement is the value from [ScanConfig.Replacements] for this
	// sequence, or an empty string if the sequence is not mapped.
	Replacement string

	// Line is the 1-based line number where the match was found.
	Line int

	// Column is the 0-based byte offset within the line.
	Column int

	// Context is the full line text containing the match.
	Context string
}

// Finding describes a file whose content differs from its sanitized form.
// Callers can inspect [Finding.HasEmoji] to determine if emoji was detected
// and write [Finding.Cleaned] back to the file to remediate it.
type Finding struct {
	// Path is the forward-slash normalized file path.
	//
	// When produced by [ScanDir], Path is relative to cfg.Root.
	// When produced by [ScanFile], Path is the argument as passed by the caller
	// (absolute or relative to the process working directory), forward-slash
	// normalized. The two sources are therefore not directly comparable unless
	// the caller passes a root-relative path to [ScanFile].
	Path string

	// HasEmoji reports whether [ContainsEmoji] detected emoji in the file.
	HasEmoji bool

	// Original is the file's content before sanitization.
	Original string

	// Cleaned is the file's content after applying [Sanitize] with the
	// configured [Options].
	Cleaned string

	// Matches holds per-occurrence detail for every emoji codepoint found in
	// Original. It is only populated when [ScanConfig.CollectMatches] is true;
	// otherwise it is nil.
	Matches []Match
}

// sniffSize is the number of bytes inspected for NUL when detecting binary files.
const sniffSize = 512

// isBinary reports whether data looks like a binary file by checking the first
// 512 bytes for a NUL byte (\x00). This mirrors the heuristic used by Git and
// other text-processing tools.
func isBinary(data []byte) bool {
	snip := data
	if len(snip) > sniffSize {
		snip = snip[:sniffSize]
	}
	return bytes.ContainsRune(snip, 0)
}

// ScanDir walks the directory tree rooted at cfg.Root and returns a [Finding]
// for every file whose content would change after applying [Sanitize] with
// cfg.Options. Files matching ExemptFiles, ExemptSuffixes, or outside the
// Extensions filter are skipped. Directories matching SkipDirs are not entered.
// Files larger than cfg.MaxFileBytes are silently skipped (zero disables the
// limit). Binary files (detected by a NUL byte in the first 512 bytes) are
// silently skipped.
//
// Unlike the pure-string functions in this package, ScanDir performs file I/O
// and therefore returns an error when the filesystem is inaccessible.
func ScanDir(cfg ScanConfig) ([]Finding, error) {
	root := cfg.Root
	if root == "" {
		root = "."
	}

	var findings []Finding

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		norm := filepath.ToSlash(rel)

		// Skip excluded directories by base name. The root (norm == ".")
		// is never skipped so that scanning always begins.
		if d.IsDir() {
			if norm != "." {
				name := d.Name()
				for _, skip := range cfg.SkipDirs {
					if name == strings.TrimSuffix(skip, "/") {
						return filepath.SkipDir
					}
				}
			}
			return nil
		}

		// Extensions filter.
		if len(cfg.Extensions) > 0 {
			matched := false
			for _, ext := range cfg.Extensions {
				if strings.HasSuffix(d.Name(), ext) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Exempt files by base name.
		base := d.Name()
		for _, exempt := range cfg.ExemptFiles {
			if base == exempt {
				return nil
			}
		}

		// Exempt files by suffix.
		for _, suffix := range cfg.ExemptSuffixes {
			if strings.HasSuffix(base, suffix) {
				return nil
			}
		}

		// MaxFileBytes size guard -- skip large/binary files.
		if cfg.MaxFileBytes > 0 {
			info, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			if info.Size() > cfg.MaxFileBytes {
				return nil
			}
		}

		// Read, sanitize, and compare.
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		// Skip binary files (NUL in first 512 bytes).
		if isBinary(data) {
			return nil
		}
		original := string(data)

		var cleaned string
		if len(cfg.Replacements) > 0 {
			cleaned = Replace(original, cfg.Replacements)
			if cfg.Options.NormalizeWhitespace {
				cleaned = Normalize(cleaned)
			}
		} else {
			cleaned = Sanitize(original, cfg.Options)
		}

		if cleaned != original {
			f := Finding{
				Path:     norm,
				HasEmoji: ContainsEmoji(original),
				Original: original,
				Cleaned:  cleaned,
			}
			if cfg.CollectMatches {
				f.Matches = buildMatches(original, cfg.Replacements)
			}
			findings = append(findings, f)
		}

		return nil
	})

	return findings, err
}

// buildMatches scans text line by line and returns a Match for every emoji
// codepoint occurrence. Each Match records the matched sequence, its 1-based
// line and 0-based byte-column within that line, the full line as context,
// and the mapped replacement string (empty if not in replacements).
//
// Replacement lookup uses the same longest-key-first greedy walk as [Replace]
// and [FindAllMapped], so a variation-selector sequence such as U+26A0 U+FE0F
// is attributed to the combined key rather than the bare codepoint when both
// appear in the map.
func buildMatches(text string, replacements map[string]string) []Match {
	keys := sortedKeys(replacements) // longest first; nil-safe (empty slice when map empty)
	var matches []Match
	for lineIdx, line := range strings.Split(text, "\n") {
		for i := 0; i < len(line); {
			// Try each replacement key longest-first so variation-selector
			// sequences (e.g., U+26A0 U+FE0F) are attributed to the combined key.
			matched := false
			for _, k := range keys {
				if strings.HasPrefix(line[i:], k) {
					matches = append(matches, Match{
						Emoji:       k,
						Replacement: replacements[k],
						Line:        lineIdx + 1,
						Column:      i,
						Context:     line,
					})
					i += len(k)
					matched = true
					break
				}
			}
			if matched {
				continue
			}
			// No replacement key matched; fall back to emojiRE for unmapped emoji.
			if loc := emojiRE.FindStringIndex(line[i:]); loc != nil && loc[0] == 0 {
				matches = append(matches, Match{
					Emoji:       line[i : i+loc[1]],
					Replacement: "",
					Line:        lineIdx + 1,
					Column:      i,
					Context:     line,
				})
				i += loc[1]
			} else {
				i++
			}
		}
	}
	return matches
}

// ScanFile checks a single file against opts and returns a [Finding] if the
// file's content would change after sanitization, or nil if the file is
// already clean. [Finding.Path] is set to path with forward slashes;
// it is not made relative to any root. Pass a root-relative path to obtain
// a [Finding] whose Path is comparable to those returned by [ScanDir].
//
// Unlike the pure-string functions in this package, ScanFile performs file I/O
// and therefore returns an error when the file is inaccessible.
func ScanFile(path string, opts Options) (*Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	original := string(data)
	cleaned := Sanitize(original, opts)
	if cleaned == original {
		return nil, nil
	}
	return &Finding{
		Path:     filepath.ToSlash(path),
		HasEmoji: ContainsEmoji(original),
		Original: original,
		Cleaned:  cleaned,
	}, nil
}
