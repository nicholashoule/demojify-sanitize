package demojify

// WriteFinding writes f.Cleaned back to the file at path atomically.
// No write occurs when f.Cleaned equals f.Original (the file is already
// clean). Original file permissions are preserved.
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
	return true, statAndWrite(path, f.Cleaned)
}
