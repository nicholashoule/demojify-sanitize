package demojify

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FixDir walks the directory tree at root, applies the sanitization or
// replacement pipeline described by cfg, and writes back every file whose
// content changed. It returns the number of files fixed (written back), the
// number of qualifying text files that were already clean (needed no
// changes), and any error from the underlying scan.
//
// The clean count includes every text file that passed all filters
// (extensions, exempt files/suffixes, size limit, binary detection) and
// whose sanitized content was identical to its original content.
//
// FixDir sets cfg.Root to root before scanning, so callers do not need to
// set it separately. All other [ScanConfig] fields -- SkipDirs, Extensions,
// ExemptFiles, ExemptSuffixes, MaxFileBytes, Options, Replacements, and
// CollectMatches -- behave identically to [ScanDir].
//
// Each resolved write target is validated to remain within root, preventing
// path-traversal writes via ".." components or symlinks in [Finding.Path].
// Both the root and each target are resolved through [filepath.EvalSymlinks]
// before comparison, so a symlink inside root that points outside it is
// also rejected.
//
// Original file permissions are preserved (see [WriteFinding]).
//
// FixDir returns an error when the directory scan fails (e.g., the root
// directory does not exist), when root cannot be resolved to an absolute
// path, or when one or more files are skipped due to path-resolution or
// write errors. In the partial-failure case the fixed and clean counts are
// still valid; callers that need per-file detail should use [ScanDir] plus
// [WriteFinding] directly.
func FixDir(root string, cfg ScanConfig) (fixed, clean int, err error) {
	cfg.Root = root

	findings, scanned, scanErr := scanDirCounted(cfg)
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
	// Resolve symlinks on root so that symlink targets inside root that
	// point outside it are caught by isInsideDir.
	realRoot, evalErr := filepath.EvalSymlinks(absRoot)
	if evalErr != nil {
		return 0, 0, fmt.Errorf("resolve root symlinks: %w", evalErr)
	}

	var skipped int
	for _, f := range findings {
		absPath := filepath.Join(root, filepath.FromSlash(f.Path))

		// Resolve to absolute, evaluate symlinks, and verify the real
		// target stays within the real root. This prevents path-traversal
		// via ".." components and via symlinks that point outside root.
		resolved, resolveErr := filepath.Abs(absPath)
		if resolveErr != nil {
			skipped++
			continue
		}
		real, evalErr := filepath.EvalSymlinks(resolved)
		if evalErr != nil {
			skipped++
			continue
		}
		if !isInsideDir(real, realRoot) {
			skipped++
			continue
		}

		changed, werr := WriteFinding(real, f)
		if werr != nil {
			skipped++
			continue
		}
		if changed {
			fixed++
		}
	}

	clean = scanned - len(findings)
	if skipped > 0 {
		return fixed, clean, fmt.Errorf("fixdir: %d file(s) skipped due to path-resolution or write errors", skipped)
	}
	return fixed, clean, nil
}

// isInsideDir reports whether target is equal to or a child of dir.
// Both paths must be absolute and cleaned (as returned by filepath.Abs).
func isInsideDir(target, dir string) bool {
	rel, err := filepath.Rel(dir, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
