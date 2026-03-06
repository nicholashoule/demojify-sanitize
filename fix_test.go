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

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}
		if clean != 1 {
			t.Errorf("clean = %d, want 1", clean)
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

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 0 {
			t.Errorf("fixed = %d, want 0", fixed)
		}
		if clean != 2 {
			t.Errorf("clean = %d, want 2", clean)
		}
	})

	t.Run("respects SkipDirs", func(t *testing.T) {
		dir := t.TempDir()
		sub := writeTempDir(t, dir, "skipme")
		writeTempFile(t, sub, "emoji.txt", "\u2705 check\n")
		writeTempFile(t, dir, "root.txt", "\u274C fail\n")

		cfg := demojify.DefaultScanConfig()
		cfg.SkipDirs = append(cfg.SkipDirs, "skipme/")

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1 (only root.txt)", fixed)
		}
		if clean != 0 {
			t.Errorf("clean = %d, want 0 (skipped dir not counted)", clean)
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

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1 (only .md)", fixed)
		}
		if clean != 0 {
			t.Errorf("clean = %d, want 0 (.txt not scanned)", clean)
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

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}
		if clean != 0 {
			t.Errorf("clean = %d, want 0", clean)
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
		fixed1, clean1, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir (1): %v", err)
		}
		if fixed1 != 1 {
			t.Errorf("first run: fixed = %d, want 1", fixed1)
		}
		if clean1 != 0 {
			t.Errorf("first run: clean = %d, want 0", clean1)
		}

		// Second run should find nothing to fix.
		fixed2, clean2, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir (2): %v", err)
		}
		if fixed2 != 0 {
			t.Errorf("second run: fixed = %d, want 0 (idempotent)", fixed2)
		}
		if clean2 != 1 {
			t.Errorf("second run: clean = %d, want 1", clean2)
		}
	})

	t.Run("multiple dirty files", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "a.txt", "\U0001F680 rocket\n")
		writeTempFile(t, dir, "b.txt", "\u2705 check\n")
		writeTempFile(t, dir, "c.txt", "clean\n")

		cfg := demojify.DefaultScanConfig()

		fixed, clean, err := demojify.FixDir(dir, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 2 {
			t.Errorf("fixed = %d, want 2", fixed)
		}
		if clean != 1 {
			t.Errorf("clean = %d, want 1", clean)
		}
	})

	t.Run("bad root returns error", func(t *testing.T) {
		badRoot := filepath.Join(t.TempDir(), "missing")
		cfg := demojify.DefaultScanConfig()

		_, _, err := demojify.FixDir(badRoot, cfg)
		if err == nil {
			t.Error("want error for nonexistent root, got nil")
		}
	})

	t.Run("fixes files in nested subdirectories", func(t *testing.T) {
		root := t.TempDir()
		sub := writeTempDir(t, root, "level1")
		deep := writeTempDir(t, sub, "level2")
		writeTempFile(t, root, "top.txt", "\U0001F680 rocket\n")
		writeTempFile(t, sub, "mid.txt", "\u2705 check\n")
		writeTempFile(t, deep, "bottom.txt", "\u274C fail\n")

		cfg := demojify.DefaultScanConfig()

		fixed, clean, err := demojify.FixDir(root, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 3 {
			t.Errorf("fixed = %d, want 3", fixed)
		}
		if clean != 0 {
			t.Errorf("clean = %d, want 0", clean)
		}

		// Verify every file on disk is now emoji-free.
		for _, rel := range []string{"top.txt", "level1/mid.txt", "level1/level2/bottom.txt"} {
			data, rerr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if rerr != nil {
				t.Fatalf("ReadFile %s: %v", rel, rerr)
			}
			if demojify.ContainsEmoji(string(data)) {
				t.Errorf("file %s still contains emoji: %q", rel, data)
			}
		}
	})

	t.Run("does not write outside root", func(t *testing.T) {
		// Integration check: FixDir only modifies files inside root.
		// A sibling temp directory with a sentinel file must remain
		// untouched.
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
		fixed, clean, err := demojify.FixDir(root, cfg)
		if err != nil {
			t.Fatalf("FixDir: %v", err)
		}
		if fixed != 1 {
			t.Errorf("fixed = %d, want 1", fixed)
		}
		if clean != 0 {
			t.Errorf("clean = %d, want 0", clean)
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

	t.Run("rejects symlink pointing outside root", func(t *testing.T) {
		// Verify FixDir's symlink-traversal protection: a symlink
		// inside root that resolves to a file outside root must not
		// be written. ScanDir follows the symlink and produces a
		// finding, but FixDir must reject the write because the real
		// target is outside the resolved root.

		root := t.TempDir()
		outside := t.TempDir()

		// Create a dirty file outside root -- the symlink target.
		target := filepath.Join(outside, "escape.txt")
		if err := os.WriteFile(target, []byte("\U0001F680 rocket\n"), 0o644); err != nil {
			t.Fatalf("write target: %v", err)
		}

		// Create a symlink inside root pointing to the outside file.
		link := filepath.Join(root, "escape.txt")
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("skipping: cannot create symlink (requires privileges on Windows): %v", err)
		}

		cfg := demojify.DefaultScanConfig()
		fixed, _, err := demojify.FixDir(root, cfg)

		// FixDir should return an error for the skipped symlink.
		if err == nil {
			t.Error("want error for symlink outside root, got nil")
		}
		if fixed != 0 {
			t.Errorf("fixed = %d, want 0 (symlink target outside root)", fixed)
		}

		// The outside file must be untouched.
		data, rerr := os.ReadFile(target)
		if rerr != nil {
			t.Fatalf("ReadFile target: %v", rerr)
		}
		if string(data) != "\U0001F680 rocket\n" {
			t.Errorf("outside file was modified: %q", data)
		}
	})
}
