package demojify

import (
	"os"
	"path/filepath"
)

// atomicWrite writes content to path using a temp-file-plus-rename strategy,
// preserving the given file permissions. On success the file at path contains
// exactly content; on failure the original file is left untouched and the
// temp file is removed.
//
// On POSIX systems rename(2) is atomic and replaces the destination in a
// single filesystem operation. On Windows, Go 1.21+ implements os.Rename via
// MoveFileEx with MOVEFILE_REPLACE_EXISTING, which replaces the destination
// but is not guaranteed atomic by the kernel. In practice this is safe for
// same-volume replace-in-place (the temp file is always created in the same
// directory).
func atomicWrite(path, content string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".demojify-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
}

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
func WriteFinding(path string, f Finding) (changed bool, err error) {
	if f.Cleaned == f.Original {
		return false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, atomicWrite(path, f.Cleaned, info.Mode().Perm())
}
