package demojify_test

import (
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestScanFile(t *testing.T) {
	opts := demojify.Options{
		RemoveEmojis: true,
	}
	tests := []struct {
		name     string
		content  string
		wantNil  bool
		wantEmoj bool
	}{
		{
			name:    "clean file returns nil",
			content: "package main\n",
			wantNil: true,
		},
		{
			name:     "file with emoji returns finding",
			content:  "package main // \U0001F680\n",
			wantNil:  false,
			wantEmoj: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			path := writeTempFile(t, root, "file.go", tt.content)

			f, err := demojify.ScanFile(path, opts)
			if err != nil {
				t.Fatalf("ScanFile: %v", err)
			}
			if tt.wantNil && f != nil {
				t.Errorf("expected nil finding, got %+v", f)
			}
			if !tt.wantNil && f == nil {
				t.Fatal("expected finding, got nil")
			}
			if !tt.wantNil && f.HasEmoji != tt.wantEmoj {
				t.Errorf("HasEmoji = %v, want %v", f.HasEmoji, tt.wantEmoj)
			}
		})
	}
}

func TestScanFileNotFound(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir", "file.go")
	_, err := demojify.ScanFile(missing, demojify.DefaultOptions())
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestScanFileSkipsBinaryFiles(t *testing.T) {
	root := t.TempDir()
	// Binary content: NUL byte in the first 512 bytes.
	binary := []byte("binary \U0001F680 content\x00 rest of file\n")
	path := writeTempFile(t, root, "image.dat", string(binary))

	f, err := demojify.ScanFile(path, demojify.Options{RemoveEmojis: true})
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if f != nil {
		t.Errorf("expected nil finding for binary file, got %+v", f)
	}
}

func TestScanFileNormalizeWhitespace(t *testing.T) {
	root := t.TempDir()
	path := writeTempFile(t, root, "messy.txt", "\U0001F680 Deploy  complete!\n\n\n\nDone.\n")

	f, err := demojify.ScanFile(path, demojify.DefaultOptions())
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	want := "Deploy complete!\n\nDone."
	if f.Cleaned != want {
		t.Errorf("Cleaned = %q, want %q", f.Cleaned, want)
	}
}

func TestScanFileEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "empty.txt", "")

	f, err := demojify.ScanFile(path, demojify.DefaultOptions())
	if err != nil {
		t.Fatalf("ScanFile on empty file: %v", err)
	}
	if f != nil {
		t.Errorf("expected nil finding for empty file, got %+v", f)
	}
}

// TestScanDirEmptyFile verifies that ScanDir handles empty (0-byte) files
// without error and does not report them as findings.
