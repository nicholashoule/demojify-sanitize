// Internal helpers shared across files: isBinary (NUL sniff), sortByLenDesc
// and sortedKeys (longest-first key ordering for the substitution scan),
// readCapped and readText (size-bounded reads), and statAndWrite/atomicWrite
// (same-directory temp file, fsync, chmod, a changed-since-read check,
// rename). sortedKeys drops empty keys; an empty key would match at every
// position.

package demojify

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ErrFileChanged is returned by the functions that write a file back
// ([WriteFinding], [FixDir], [SanitizeFile], [ReplaceFile]) when the file no
// longer holds the content that was read and cleaned: another writer changed
// it in between. Nothing is written; the file keeps the other writer's
// content. Branch on it with [errors.Is].
//
// The detection is best-effort. The file is compared once the new content
// is ready, just before the rename; a write that lands between that check
// and the rename is not seen, and the rename replaces it. No portable
// filesystem operation compares and renames atomically. Treat it as
// stale-write detection, not a lock: callers whose writers must exclude one
// another coordinate them.
var ErrFileChanged = errors.New("demojify: file changed since it was read")

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
// codepoint). Empty keys are silently omitted because an empty key would
// prefix-match at every byte position in the substitution scan, emitting its
// replacement between every rune and causing unbounded output growth -- a
// potential denial-of-service vector for any caller that passes
// user-supplied map keys.
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

// readCapped reads r to EOF, at most limit bytes. over reports that r held
// more than limit bytes, in which case data is nil: the caller skips the
// input instead of reading it whole.
func readCapped(r io.Reader, limit int64) (data []byte, over bool, err error) {
	data, err = io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		return nil, true, nil
	}
	return data, false, nil
}

// readText reads a file's content from r for the [ScanDir] walk. It sniffs
// the first [sniffSize] bytes for a NUL byte before reading the rest, so a
// large binary file costs one small read. With limit > 0 the whole read,
// sniff included, is at most limit+1 bytes: the sniff itself shrinks for a
// limit under sniffSize, so even a small limit is never read past by more
// than the one byte that shows the input is over it. skip reports a binary
// file or one over the limit; data is nil then.
func readText(r io.Reader, limit int64) (data []byte, skip bool, err error) {
	var sniff [sniffSize]byte
	buf := sniff[:]
	if limit > 0 && limit < sniffSize {
		buf = sniff[:limit+1]
	}
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, false, err
	}
	if (limit > 0 && int64(n) > limit) || isBinary(sniff[:n]) {
		return nil, true, nil
	}
	var rest []byte
	if limit > 0 {
		var over bool
		if rest, over, err = readCapped(r, limit-int64(n)); err == nil && over {
			return nil, true, nil
		}
	} else {
		rest, err = io.ReadAll(r)
	}
	if err != nil {
		return nil, false, err
	}
	data = make([]byte, n+len(rest))
	copy(data, sniff[:n])
	copy(data[n:], rest)
	return data, false, nil
}

// statAndWrite writes cleaned through a same-directory temp file and rename,
// preserving the file's current permissions, provided the file still holds
// original (the content cleaned was derived from); otherwise it returns an
// error wrapping [ErrFileChanged] and writes nothing. It is used by
// file-modifying functions after they have determined that a write is
// necessary.
func statAndWrite(path, original, cleaned string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return atomicWrite(path, cleaned, info.Mode().Perm(), original)
}

// unchanged reports whether the file at path still holds exactly want. It
// reads at most len(want)+1 bytes, so a file that grew is caught without
// being read whole.
func unchanged(path, want string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()
	got, over, err := readCapped(f, int64(len(want)))
	if err != nil {
		return false, err
	}
	return !over && string(got) == want, nil
}

// atomicWrite writes content to path using a temp-file-plus-rename strategy,
// preserving the given file permissions. On success the file at path contains
// exactly content; on failure the original file is left untouched and the
// temp file is removed.
//
// The rename happens only if the file at path still holds original, checked
// once the temp file is complete so the window between the check and the
// rename is as short as the filesystem allows. A file another writer changed
// after it was read, up to that check, is reported as [ErrFileChanged]
// instead of being overwritten with content cleaned from the stale copy. The
// window itself stays open (see [ErrFileChanged]).
//
// On POSIX systems rename(2) is atomic and replaces the destination in a
// single filesystem operation. On Windows, Go (1.24+, the module minimum) implements os.Rename via
// MoveFileEx with MOVEFILE_REPLACE_EXISTING, which replaces the destination
// but is not guaranteed atomic by the kernel. In practice this is safe for
// same-volume replace-in-place (the temp file is always created in the same
// directory). The file data is flushed to stable storage via Sync before
// the rename so that a power failure cannot leave a zero-length or
// partially-written destination.
func atomicWrite(path, content string, perm os.FileMode, original string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".demojify-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	same, err := unchanged(path, original)
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("%w: %s", ErrFileChanged, path)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
}
