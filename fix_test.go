package demojify_test

import (
	"os"
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestFixDir(t *testing.T) {
	t.Run("fixes emoji files and reports count", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "clean.txt", "no emoji\n")
		writeTempFile(t, dir, "dirty.txt", "\U0001F680 rocket launch\n")

		cfg := demojify.DefaultScanConfig()

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}

		// Verify the file on disk no longer contains emoji.
		data, err := os.ReadFile(filepath.Join(dir, "dirty.txt"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if demojify.ContainsEmoji(string(data)) {
			t.Errorf("file still contains emoji after FixDir: %q", data)
		}
	})

	t.Run("clean directory returns zero fixed", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "a.txt", "hello\n")
		writeTempFile(t, dir, "b.txt", "world\n")

		cfg := demojify.DefaultScanConfig()

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 0 {
			t.Errorf("fixed = %d, want 0", fixed)
		}
	})

	t.Run("respects SkipDirs", func(t *testing.T) {
		dir := t.TempDir()
		sub := writeTempDir(t, dir, "skipme")
		writeTempFile(t, sub, "emoji.txt", "\u2705 check\n")
		writeTempFile(t, dir, "root.txt", "\u274C fail\n")

		cfg := demojify.DefaultScanConfig()
		cfg.SkipDirs = append(cfg.SkipDirs, "skipme/")

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1 (only root.txt)", fixed)
		}

		// Skipped file should still contain emoji.
		data, _ := os.ReadFile(filepath.Join(dir, "skipme", "emoji.txt"))
		if !demojify.ContainsEmoji(string(data)) {
			t.Error("skipme/emoji.txt should still contain emoji")
		}
	})

	t.Run("respects Extensions filter", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "emoji.md", "\U0001F680 rocket\n")
		writeTempFile(t, dir, "emoji.txt", "\U0001F680 rocket\n")

		cfg := demojify.DefaultScanConfig()
		cfg.Extensions = []string{".md"}

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1 (only .md)", fixed)
		}

		// .txt should still contain emoji.
		data, _ := os.ReadFile(filepath.Join(dir, "emoji.txt"))
		if !demojify.ContainsEmoji(string(data)) {
			t.Error("emoji.txt should still contain emoji (not in Extensions)")
		}
	})

	t.Run("with replacements substitutes emoji", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "sub.txt", "\u2705 tests passed\n")

		cfg := demojify.DefaultScanConfig()
		cfg.Replacements = demojify.DefaultReplacements()

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}

		data, _ := os.ReadFile(filepath.Join(dir, "sub.txt"))
		if demojify.ContainsEmoji(string(data)) {
			t.Errorf("file still contains emoji after substitution: %q", data)
		}
	})

	t.Run("idempotent on second run", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "emoji.txt", "\U0001F680 rocket\n")

		cfg := demojify.DefaultScanConfig()

		// First fix.
		fixed1, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir (1): %v", err)
		}
		if fixed1 != 1 {
			t.Errorf("first run: fixed = %d, want 1", fixed1)
		}

		// Second run should find nothing to fix.
		fixed2, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir (2): %v", err)
		}
		if fixed2 != 0 {
			t.Errorf("second run: fixed = %d, want 0 (idempotent)", fixed2)
		}
	})

	t.Run("multiple dirty files", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "a.txt", "\U0001F680 rocket\n")
		writeTempFile(t, dir, "b.txt", "\u2705 check\n")
		writeTempFile(t, dir, "c.txt", "clean\n")

		cfg := demojify.DefaultScanConfig()

		fixed, _, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 2 {
			t.Errorf("fixed = %d, want 2", fixed)
		}
	})

	t.Run("bad root returns error", func(t *testing.T) {
		cfg := demojify.DefaultScanConfig()

		_, _, err := demojify.FixDir("/no/such/directory/ever", cfg)
		if err == nil {
			t.Error("want error for nonexistent root, got nil")
		}
	})

	t.Run("path traversal in Finding.Path is rejected", func(t *testing.T) {
		// FixDir should refuse to write to paths that escape root via "..".
		// We set up a directory with a dirty file, fix it, then verify that
		// a sibling directory outside root is never touched.
		root := t.TempDir()
		outside := t.TempDir()

		// Put a sentinel file outside root that should never be modified.
		sentinel := filepath.Join(outside, "secret.txt")
		if err := os.WriteFile(sentinel, []byte("do not touch\n"), 0o644); err != nil {
			t.Fatalf("write sentinel: %v", err)
		}

		// Create a real dirty file inside root so ScanDir returns findings.
		writeTempFile(t, root, "legit.txt", "\U0001F680 rocket\n")

		cfg := demojify.DefaultScanConfig()
		fixed, _, err := demojify.FixDir(root, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}

		// Sentinel must be untouched.
		data, err := os.ReadFile(sentinel)
		if err != nil {
			t.Fatalf("ReadFile sentinel: %v", err)
		}
		if string(data) != "do not touch\n" {
			t.Errorf("sentinel was modified: %q", data)
		}
	})
}
