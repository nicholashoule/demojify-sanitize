// WriteFinding: permission-preserving write-back of a Finding via statAndWrite,
// refused when the file changed after it was scanned.

package demojify

// WriteFinding writes f.Cleaned back through a same-directory temp file and
// rename (atomic on POSIX, best-effort replacement on Windows). No write
// occurs when f.Cleaned equals f.Original (the file is already clean).
// Original file permissions are preserved.
//
// The write happens only if the file still holds f.Original, the content the
// scan read and cleaned. A file another writer changed after the scan is
// left alone and WriteFinding returns (false, err) with err wrapping
// [ErrFileChanged], so a stale [Finding] never overwrites newer content.
//
// WriteFinding returns true when the file was modified and false when it
// was already clean. It returns an error for any filesystem failure.
//
// The path argument must be the resolved filesystem path to the file.
// When used with [ScanDir] results, join [ScanConfig.Root] with
// [Finding.Path] to produce the correct path:
//
//	absPath := filepath.Join(cfg.Root, filepath.FromSlash(f.Path))
//	changed, err := demojify.WriteFinding(absPath, f)
//
// When used with [ScanFile] results the returned [Finding.Path] can
// typically be passed directly, since it matches the original argument.
func WriteFinding(path string, f Finding) (changed bool, err error) { //nolint:gocritic // hugeParam: Finding is passed by value per public API contract
	if f.Cleaned == f.Original {
		return false, nil
	}
	if err := statAndWrite(path, f.Original, f.Cleaned); err != nil {
		return false, err
	}
	return true, nil
}
