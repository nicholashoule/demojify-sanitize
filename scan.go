package demojify

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// buildReplacer constructs a *strings.Replacer from pre-sorted keys and the
// replacements map. Called once before a directory walk to avoid rebuilding
// the replacer for every file.
func buildReplacer(keys []string, replacements map[string]string) *strings.Replacer {
	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k, replacements[k])
	}
	return strings.NewReplacer(args...)
}

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
	// When non-empty, [ScanDir] cleans files with [Replace] instead of [Sanitize],
	// so matched sequences are substituted before any residual emoji are stripped.
	// Longer keys are matched before shorter ones (variation-selector aware).
	// Has no effect on [ScanFile], which accepts only [Options].
	//
	// NOTE: When Replacements is non-empty, [Options.RemoveEmojis],
	// [Options.AllowedRanges], and [Options.AllowedEmojis] are ignored
	// because [Replace] always strips residual emoji via [Demojify]. Only
	// [Options.NormalizeWhitespace] is honored after substitution; when
	// enabled it runs unconditionally on every scanned file. To
	// preserve specific Unicode ranges during replacement-based scans, add
	// those codepoints to the Replacements map with identity values
	// (key == value).
	Replacements map[string]string

	// CollectMatches controls whether [ScanDir] populates [Finding.Matches] for
	// each file whose content differs from its sanitized form. When true, every
	// matched codepoint sequence -- replacement keys (which may include
	// non-emoji characters) and unmapped emoji -- is recorded with its line,
	// column, and surrounding context. Setting this flag carries a small
	// per-file allocation cost; leave it false for bulk audits that only need
	// the cleaned text.
	CollectMatches bool
}

// DefaultScanConfig returns a ScanConfig suitable for auditing a typical Go
// module repository or AI-agent workspace. It skips .git/, vendor/, and
// node_modules/ directories, exempts *_test.go files, and scans all file
// types by default (Extensions is nil). Files larger than 1 MiB are skipped
// to avoid reading large files into memory, and binary files (detected by a
// NUL byte in the first 512 bytes) are always skipped regardless of size.
//
// NOTE for downstream consumers: the default config exempts *_test.go files
// via ExemptSuffixes. This is appropriate for scanning this module's own repo
// (where test files intentionally contain literal emoji as input data), but
// downstream projects that want to scan their own test files should clear or
// override ExemptSuffixes:
//
//	cfg := demojify.DefaultScanConfig()
//	cfg.ExemptSuffixes = nil // scan test files too
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

// Match describes a single matched codepoint sequence within a scanned file.
// A match may be an emoji detected by the internal regex, or any mapped
// sequence from [ScanConfig.Replacements] (which may include non-emoji
// codepoints such as arrows or geometric shapes). When Replacements is
// empty, only emoji codepoints are recorded.
// It is populated in [Finding.Matches] when [ScanConfig.CollectMatches] is true.
type Match struct {
	// Sequence is the matched codepoint sequence. This may hold non-emoji
	// sequences (such as arrows or geometric shapes) when
	// [ScanConfig.Replacements] maps those codepoints.
	Sequence string

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

	// Matches holds per-occurrence detail for every matched codepoint sequence
	// found in Original. When [ScanConfig.Replacements] is non-empty, this includes
	// both replacement keys (which may be non-emoji) and unmapped emoji.
	// It is only populated when [ScanConfig.CollectMatches] is true;
	// otherwise it is nil.
	Matches []Match
}

// ScanDir walks the directory tree rooted at cfg.Root and returns a [Finding]
// for every file whose content would change after cleaning. When
// [ScanConfig.Replacements] is non-empty, each file is cleaned with [Replace]
// (mapped-sequence substitution followed by residual-emoji stripping via
// [Demojify]); otherwise emoji removal is applied per cfg.Options.
//
// When [Options.NormalizeWhitespace] is enabled, whitespace normalization
// is applied unconditionally to every scanned file, regardless of whether
// emoji were found or replaced. This guarantees that redundant whitespace
// is always cleaned in a single pass.
//
// Files matching ExemptFiles, ExemptSuffixes, or outside the
// Extensions filter are skipped. Directories matching SkipDirs are not entered.
// Files larger than cfg.MaxFileBytes are silently skipped (zero disables the
// limit). Binary files (detected by a NUL byte in the first 512 bytes) are
// silently skipped.
//
// Unlike the pure-string functions in this package, ScanDir performs file I/O
// and therefore returns an error when the filesystem is inaccessible.
func ScanDir(cfg ScanConfig) ([]Finding, error) { //nolint:gocritic // hugeParam: ScanConfig is passed by value per public API contract
	findings, _, err := scanDirCounted(context.Background(), cfg)
	return findings, err
}

// ScanDirContext is like [ScanDir] but accepts a [context.Context] for
// cancellation support. When the context is canceled the directory walk
// stops and ScanDirContext returns the context error along with any
// findings collected before cancellation. Agent orchestrators and MCP
// servers can use this to enforce timeouts on large repository scans.
func ScanDirContext(ctx context.Context, cfg ScanConfig) ([]Finding, error) { //nolint:gocritic // hugeParam: ScanConfig is passed by value per public API contract
	findings, _, err := scanDirCounted(ctx, cfg)
	return findings, err
}

// scanDirCounted is the internal implementation of [ScanDir]. It returns
// the same findings slice plus a count of every qualifying text file that
// was scanned (whether or not its content changed). Callers that need the
// total file count -- such as [FixDir] -- use this directly.
func scanDirCounted(ctx context.Context, cfg ScanConfig) ([]Finding, int, error) { //nolint:gocritic // hugeParam: mirrors public API callers
	root := cfg.Root
	if root == "" {
		root = "."
	}

	// Precompute trimmed skip-dir names so strings.TrimSuffix is not
	// called on every directory entry during the walk.
	trimmedSkips := make([]string, len(cfg.SkipDirs))
	for i, s := range cfg.SkipDirs {
		trimmedSkips[i] = strings.TrimSuffix(s, "/")
	}

	// Pre-sort replacement keys and build the replacer once so the walk
	// callback does not re-sort and re-allocate for every file.
	var replKeys []string
	var replacer *strings.Replacer
	if len(cfg.Replacements) > 0 {
		replKeys = sortedKeys(cfg.Replacements)
		replacer = buildReplacer(replKeys, cfg.Replacements)
	}

	var findings []Finding
	var scanned int

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		// Check for context cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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
				for _, skip := range trimmedSkips {
					if name == skip {
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

		// MaxFileBytes size guard -- skip large files.
		if cfg.MaxFileBytes > 0 {
			info, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			if info.Size() > cfg.MaxFileBytes {
				return nil
			}
		}

		// Two-phase read: sniff the first 512 bytes for a NUL byte
		// to detect binary files before committing to a full read.
		// This avoids unnecessary I/O and allocations for large
		// binary files, especially when MaxFileBytes is 0 (disabled).
		f, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		var sniff [sniffSize]byte
		n, sniffErr := io.ReadFull(f, sniff[:])
		if sniffErr != nil && sniffErr != io.ErrUnexpectedEOF && sniffErr != io.EOF {
			f.Close()
			return sniffErr
		}
		if isBinary(sniff[:n]) {
			f.Close()
			return nil
		}
		// Text file -- read the remainder and combine.
		rest, readErr := io.ReadAll(f)
		f.Close()
		if readErr != nil {
			return readErr
		}
		var data []byte
		if len(rest) > 0 {
			data = make([]byte, n+len(rest))
			copy(data, sniff[:n])
			copy(data[n:], rest)
		} else {
			data = sniff[:n]
		}
		original := string(data)

		// Count this file as scanned -- it passed all filters and is a
		// qualifying text file.
		scanned++

		// Clean the file content. Emoji removal (or replacement-based
		// substitution) runs first. Whitespace normalization, when enabled,
		// runs unconditionally on the result.
		var cleaned string
		//nolint:gocritic // ifElseChain: switch would require nesting a second switch for AllowedEmojis/AllowedRanges
		if replacer != nil {
			cleaned = Demojify(replacer.Replace(original))
		} else if cfg.Options.RemoveEmojis {
			switch {
			case len(cfg.Options.AllowedEmojis) > 0:
				cleaned = demojifyPreserving(original, cfg.Options.AllowedEmojis, cfg.Options.AllowedRanges)
			case len(cfg.Options.AllowedRanges) > 0:
				cleaned = demojifyAllowed(original, cfg.Options.AllowedRanges)
			default:
				cleaned = Demojify(original)
			}
		} else {
			cleaned = original
		}

		if cfg.Options.NormalizeWhitespace {
			cleaned = Normalize(cleaned)
		}

		if cleaned != original {
			f := Finding{
				Path:     norm,
				HasEmoji: ContainsEmoji(original),
				Original: original,
				Cleaned:  cleaned,
			}
			if cfg.CollectMatches {
				f.Matches = buildMatches(original, cfg.Replacements, replKeys)
			}
			findings = append(findings, f)
		}

		return nil
	})

	return findings, scanned, err
}

// buildMatches scans text line by line and returns a Match for every matched
// codepoint sequence. When replacements is non-empty, replacement keys are
// matched first using the same longest-key-first greedy walk as [Replace]
// and [FindAllMapped]; remaining unmapped emoji are then matched via emojiRE.
// Replacement keys may include non-emoji sequences (such as arrows or
// geometric shapes); those are recorded alongside emoji matches.
//
// Each Match records the matched sequence, its 1-based line and 0-based
// byte-column within that line, the full line as context, and the mapped
// replacement string (empty if not in replacements).
//
// The keys parameter must be the replacement map keys sorted by descending
// byte length (longest first). Callers that have already sorted keys (e.g.,
// [ScanDir]) pass them directly to avoid re-sorting per file.
func buildMatches(text string, replacements map[string]string, keys []string) []Match {
	var matches []Match
	for lineIdx, line := range strings.Split(text, "\n") {
		for i := 0; i < len(line); {
			// Try each replacement key longest-first so variation-selector
			// sequences (e.g., U+26A0 U+FE0F) are attributed to the combined key.
			matched := false
			for _, k := range keys {
				if strings.HasPrefix(line[i:], k) {
					matches = append(matches, Match{
						Sequence: k,

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
			if loc := emojiRE.FindStringIndex(line[i:]); len(loc) > 0 && loc[0] == 0 {
				seq := line[i : i+loc[1]]
				matches = append(matches, Match{
					Sequence: seq,

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
// already clean. Binary files (detected by a NUL byte in the first 512 bytes)
// are silently skipped and return (nil, nil), matching the behavior of
// [ScanDir]. [Finding.Path] is set to path with forward slashes; it is not
// made relative to any root. Pass a root-relative path to obtain a [Finding]
// whose Path is comparable to those returned by [ScanDir].
//
// Unlike the pure-string functions in this package, ScanFile performs file I/O
// and therefore returns an error when the file is inaccessible.
func ScanFile(path string, opts Options) (*Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isBinary(data) {
		return nil, nil
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

// FindMatchesInFile reads the file at path and returns a [Match] for every
// matched sequence found, with [Match.Replacement] populated from the
// replacements map (empty string if the sequence is not mapped). Matches
// are ordered by line then column.
//
// Matching uses two passes per line: replacement keys are tried first
// (longest-first, so variation-selector sequences are consumed atomically),
// then any remaining codepoints are checked against the internal emoji regex.
// Because replacement keys may include non-emoji sequences (e.g., arrows),
// the returned matches are not limited to emoji codepoints. Returns nil and
// no error when the file contains no matched sequences.
//
// Binary files (detected by a NUL byte in the first 512 bytes) are silently
// skipped and return (nil, nil), matching the behavior of [ScanDir] and
// [ScanFile].
//
// Unlike [ScanDir] with CollectMatches, this function does not filter or
// sanitize the file; it only collects match metadata.
//
// FindMatchesInFile returns an error for any filesystem failure.
func FindMatchesInFile(path string, replacements map[string]string) ([]Match, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isBinary(data) {
		return nil, nil
	}
	return buildMatches(string(data), replacements, sortedKeys(replacements)), nil
}
