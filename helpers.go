package demojify

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
)

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
	return bytes.IndexByte(snip, 0) >= 0
}

// sortByLenDesc sorts a string slice in place by descending byte length
// so that longer sequences are matched before shorter sub-sequences.
func sortByLenDesc(s []string) {
	sort.Slice(s, func(i, j int) bool {
		return len(s[i]) > len(s[j])
	})
}

// sortedKeys returns the keys of m sorted by descending byte length so that
// longer sequences are matched before shorter sub-sequences (e.g., a key
// containing a variation selector such as U+FE0F is tried before its base
// codepoint). Empty keys are silently omitted because an empty key in
// [strings.NewReplacer] inserts a replacement between every rune, causing
// unbounded memory growth -- a potential denial-of-service vector for any
// caller that passes user-supplied map keys.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			keys = append(keys, k)
		}
	}
	sortByLenDesc(keys)
	return keys
}

// statAndWrite writes cleaned to path atomically, preserving the file's
// current permissions. It is used by file-modifying functions after they
// have already determined that a write is necessary (cleaned != original).
func statAndWrite(path, cleaned string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return atomicWrite(path, cleaned, info.Mode().Perm())
}

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
// directory). The file data is flushed to stable storage via Sync before
// the rename so that a power failure cannot leave a zero-length or
// partially-written destination.
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
	if err := tmp.Sync(); err != nil {
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
