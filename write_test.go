package demojify_test

import (
	"os"
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestWriteFinding(t *testing.T) {
	t.Run("dirty finding is written back", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "dirty.txt")
		original := "build \U0001F680 deployed\n"
		if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		f := demojify.Finding{
			Path:     "dirty.txt",
			HasEmoji: true,
			Original: original,
			Cleaned:  "build  deployed\n",
		}
		changed, err := demojify.WriteFinding(path, f)
		if err != nil {
			t.Fatalf("WriteFinding: %v", err)
		}
		if !changed {
			t.Error("changed = false, want true")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != f.Cleaned {
			t.Errorf("file content = %q, want %q", string(data), f.Cleaned)
		}
	})

	t.Run("clean finding is not written", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "clean.txt")
		content := "no emoji here\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		info1, _ := os.Stat(path)

		f := demojify.Finding{
			Path:     "clean.txt",
			Original: content,
			Cleaned:  content,
		}
		changed, err := demojify.WriteFinding(path, f)
		if err != nil {
			t.Fatalf("WriteFinding: %v", err)
		}
		if changed {
			t.Error("changed = true, want false for clean finding")
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
		infoBefore, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat before: %v", err)
		}
		wantPerm := infoBefore.Mode().Perm()

		f := demojify.Finding{
			Path:     "perms.txt",
			HasEmoji: true,
			Original: "check \u2705\n",
			Cleaned:  "check \n",
		}
		if _, err := demojify.WriteFinding(path, f); err != nil {
			t.Fatalf("WriteFinding: %v", err)
		}
		infoAfter, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat after: %v", err)
		}
		if infoAfter.Mode().Perm() != wantPerm {
			t.Errorf("permissions changed: got %o, want %o", infoAfter.Mode().Perm(), wantPerm)
		}
	})

	t.Run("nonexistent path returns error", func(t *testing.T) {
		f := demojify.Finding{
			Path:     "ghost.txt",
			Original: "old",
			Cleaned:  "new",
		}
		_, err := demojify.WriteFinding("/nonexistent/path/ghost.txt", f)
		if err == nil {
			t.Error("expected error for nonexistent path, got nil")
		}
	})

	t.Run("ScanDir + WriteFinding integration avoids double read", func(t *testing.T) {
		root := t.TempDir()
		writeTempFile(t, root, "status.md", "\u2705 Success\n")
		writeTempFile(t, root, "clean.md", "No emoji\n")

		cfg := demojify.DefaultScanConfig()
		cfg.Root = root
		cfg.Extensions = []string{".md"}

		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1", len(findings))
		}

		f := findings[0]
		absPath := filepath.Join(root, filepath.FromSlash(f.Path))
		changed, err := demojify.WriteFinding(absPath, f)
		if err != nil {
			t.Fatalf("WriteFinding: %v", err)
		}
		if !changed {
			t.Error("WriteFinding changed = false, want true")
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != f.Cleaned {
			t.Errorf("file = %q, want %q", string(data), f.Cleaned)
		}
	})
}
