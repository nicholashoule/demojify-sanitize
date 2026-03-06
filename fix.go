package demojify

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FixDir walks the directory tree at root, applies the sanitisation or
// replacement pipeline described by cfg, and writes back every file whose
// content changed. It returns the number of files fixed, the number that
// were already clean, and any error from the underlying [ScanDir] call.
//
// FixDir sets cfg.Root to root before scanning, so callers do not need to
// set it separately. All other [ScanConfig] fields -- SkipDirs, Extensions,
// ExemptFiles, ExemptSuffixes, MaxFileBytes, Options, Replacements, and
// CollectMatches -- behave identically to [ScanDir].
//
// Each resolved write target is validated to remain within root, preventing
// path-traversal writes if a [Finding.Path] were ever crafted with ".."
// components. Files that fail validation or write are silently skipped; use
// [ScanDir] plus [WriteFinding] directly if per-file error handling is
// required.
//
// Original file permissions are preserved (see [WriteFinding]).
//
// FixDir returns an error only when [ScanDir] itself fails (e.g., the root
// directory does not exist).
func FixDir(root string, cfg ScanConfig) (fixed, clean int, err error) {
	cfg.Root = root

	findings, scanErr := ScanDir(cfg)
	if scanErr != nil {
		return 0, 0, scanErr
	}

	if root == "" {
		root = "."
	}

	absRoot, absErr := filepath.Abs(root)
	if absErr != nil {
		return 0, 0, fmt.Errorf("resolve root: %w", absErr)
	}

	for _, f := range findings {
		absPath := filepath.Join(root, filepath.FromSlash(f.Path))

		// Resolve to absolute and verify the target stays within root.
		// This prevents path-traversal via ".." in Finding.Path.
		resolved, resolveErr := filepath.Abs(absPath)
		if resolveErr != nil {
			continue
		}
		if !isInsideDir(resolved, absRoot) {
			continue
		}

		changed, werr := WriteFinding(absPath, f)
		if werr != nil {
			continue
		}
		if changed {
			fixed++
		}
	}

	return fixed, 0, nil
}

// isInsideDir reports whether target is equal to or a child of dir.
// Both paths must be absolute and cleaned (as returned by filepath.Abs).
func isInsideDir(target, dir string) bool {
	// Ensure dir ends with a separator so that "/tmp/rootExtra" is not
	// considered inside "/tmp/root".
	prefix := dir
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return target == dir || strings.HasPrefix(target, prefix)
}
