package demojify_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestReplaceFile(t *testing.T) {
	t.Run("file with mapped emoji is substituted and written back", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "file.txt", "build \u2705 passed\n")
		repl := map[string]string{"\u2705": "[PASS]"}
		count, err := demojify.ReplaceFile(path, repl)
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count == 0 {
			t.Error("count = 0, want > 0")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile after ReplaceFile: %v", err)
		}
		want := "build [PASS] passed\n"
		if string(data) != want {
			t.Errorf("file content = %q, want %q", string(data), want)
		}
	})

	t.Run("clean file returns zero count and is not written", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "clean.txt", "no emoji here\n")
		// Record mtime before the call.
		info1, _ := os.Stat(path)
		count, err := demojify.ReplaceFile(path, map[string]string{"\u2705": "[PASS]"})
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
		info2, _ := os.Stat(path)
		if !info2.ModTime().Equal(info1.ModTime()) {
			t.Error("clean file was modified unexpectedly")
		}
	})

	t.Run("file permissions are preserved", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "perms.txt")
		if err := os.WriteFile(path, []byte("check \u2705\n"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		// Record permissions as set by the OS (Windows may normalise 0600 -> 0666).
		infoBefore, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat before: %v", err)
		}
		wantPerm := infoBefore.Mode().Perm()
		if _, err := demojify.ReplaceFile(path, map[string]string{"\u2705": "[PASS]"}); err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		infoAfter, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat after: %v", err)
		}
		if infoAfter.Mode().Perm() != wantPerm {
			t.Errorf("permissions changed: got %o, want %o", infoAfter.Mode().Perm(), wantPerm)
		}
	})

	t.Run("nil map behaves like Demojify on file", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "demojify.txt", "rocket \U0001F680\n")
		count, err := demojify.ReplaceFile(path, nil)
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count == 0 {
			t.Error("count = 0, want > 0")
		}
		data, _ := os.ReadFile(path)
		if string(data) != "rocket \n" {
			t.Errorf("file content = %q, want %q", string(data), "rocket \n")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir", "no-file.txt")
		_, err := demojify.ReplaceFile(missing, nil)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("binary file is skipped", func(t *testing.T) {
		dir := t.TempDir()
		// Embed a NUL byte in the first 512 bytes so isBinary returns true.
		path := writeTempFile(t, dir, "data.bin", "hello\x00\u2705 world\n")
		count, err := demojify.ReplaceFile(path, map[string]string{"\u2705": "[PASS]"})
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0 for binary file", count)
		}
		// Verify the file was not modified.
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != "hello\x00\u2705 world\n" {
			t.Errorf("binary file was modified: %q", string(data))
		}
	})

	t.Run("CRLF file: emoji replaced, CR bytes pass through", func(t *testing.T) {
		// Replace does not modify non-emoji bytes, so CR from CRLF endings must
		// survive the round-trip unchanged.
		dir := t.TempDir()
		path := writeTempFile(t, dir, "crlf.txt", "\u2705 passed\r\n\u274c failed\r\n")
		count, err := demojify.ReplaceFile(path, map[string]string{
			"\u2705": "[PASS]",
			"\u274c": "[FAIL]",
		})
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		want := "[PASS] passed\r\n[FAIL] failed\r\n"
		if string(data) != want {
			t.Errorf("file = %q, want %q", string(data), want)
		}
	})
}

func TestReplaceConcurrent(t *testing.T) {
	const goroutines = 50
	repl := demojify.DefaultReplacements()
	inputs := []string{
		"\u2705 build passed \u274c deploy failed",
		"\U0001F680 rocket launch \U0001F4CA chart",
		"no emoji at all",
		"\u26a0\ufe0f warning sign",
		"",
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			input := inputs[idx%len(inputs)]
			_ = demojify.Replace(input, repl)
			_, _ = demojify.ReplaceCount(input, repl)
			_ = demojify.FindAllMapped(input, repl)
		}(i)
	}
	wg.Wait()
}
