package demojify

import (
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

	// Options configures the sanitization pipeline applied to each file's
	// content. See [Options] and [DefaultOptions].
	Options Options
}

// DefaultScanConfig returns a ScanConfig suitable for auditing a typical Go
// module repository or AI-agent workspace. It skips .git/, vendor/, and
// node_modules/ directories, exempts *_test.go files, and scans all file
// types by default (Extensions is nil).
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
		Extensions:     nil, // scan all file types
		Options: Options{
			RemoveEmojis: true,
		},
	}
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
}

// ScanDir walks the directory tree rooted at cfg.Root and returns a [Finding]
// for every file whose content would change after applying [Sanitize] with
// cfg.Options. Files matching ExemptFiles, ExemptSuffixes, or outside the
// Extensions filter are skipped. Directories matching SkipDirs are not entered.
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

		// Skip excluded directories.
		if d.IsDir() {
			for _, skip := range cfg.SkipDirs {
				dir := strings.TrimSuffix(skip, "/")
				if norm == dir ||
					strings.HasPrefix(norm, dir+"/") ||
					strings.HasSuffix(norm, "/"+dir) ||
					strings.Contains(norm, "/"+dir+"/") {
					return filepath.SkipDir
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

		// Read, sanitize, and compare.
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		original := string(data)
		cleaned := Sanitize(original, cfg.Options)

		if cleaned != original {
			findings = append(findings, Finding{
				Path:     norm,
				HasEmoji: ContainsEmoji(original),
				Original: original,
				Cleaned:  cleaned,
			})
		}

		return nil
	})

	return findings, err
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
