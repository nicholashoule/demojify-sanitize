package demojify_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDefaultScanConfig(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	if cfg.Root != "." {
		t.Errorf("Root = %q, want %q", cfg.Root, ".")
	}
	if len(cfg.SkipDirs) != 3 {
		t.Errorf("SkipDirs length = %d, want 3", len(cfg.SkipDirs))
	}
	if len(cfg.ExemptSuffixes) != 1 || cfg.ExemptSuffixes[0] != "_test.go" {
		t.Errorf("ExemptSuffixes = %v, want [_test.go]", cfg.ExemptSuffixes)
	}
	if cfg.Extensions != nil {
		t.Errorf("Extensions = %v, want nil (scan all file types by default)", cfg.Extensions)
	}
	if !cfg.Options.RemoveEmojis {
		t.Error("DefaultScanConfig Options should enable emoji removal")
	}
	if cfg.Options.NormalizeWhitespace {
		t.Error("DefaultScanConfig should not enable NormalizeWhitespace (file formatters own trailing-newline conventions)")
	}
	if cfg.MaxFileBytes != 1<<20 {
		t.Errorf("MaxFileBytes = %d, want %d (1 MiB)", cfg.MaxFileBytes, 1<<20)
	}
}

func TestScanDirMaxFileBytes(t *testing.T) {
	root := t.TempDir()
	// small file (5 bytes: 4-byte emoji + newline) with emoji -- should be found
	writeTempFile(t, root, "small.go", "\U0001F680\n")
	// large file (10 bytes) with emoji -- should be skipped
	large := make([]byte, 10)
	copy(large, "\U0001F680\n")
	writeTempFile(t, root, "large.go", string(large))

	cfg := demojify.ScanConfig{
		Root:         root,
		MaxFileBytes: 8, // small.go (5 bytes) passes; large.go (10 bytes) is skipped
		Options:      demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Fatalf("got %d findings %v, want 1 (small.go only)", len(findings), paths)
	}
	if findings[0].Path != "small.go" {
		t.Errorf("finding path = %q, want \"small.go\"", findings[0].Path)
	}
}

func TestScanDirMaxFileBytesZeroMeansNoLimit(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "a.go", "package p // \U0001F680\n")
	writeTempFile(t, root, "b.go", "package p // \U0001F680\n")

	cfg := demojify.ScanConfig{
		Root:         root,
		MaxFileBytes: 0, // zero = no limit
		Options:      demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 2 {
		t.Errorf("got %d findings, want 2 (no size limit)", len(findings))
	}
}

func TestScanDirSkipsBinaryFiles(t *testing.T) {
	root := t.TempDir()

	// Text file with emoji -- should produce a finding.
	writeTempFile(t, root, "text.go", "package p // \U0001F680\n")

	// Binary file: has a NUL byte in the first 512 bytes.
	// Even though it contains an emoji codepoint, it should be skipped.
	binary := []byte("binary \U0001F680 content\x00 rest of file\n")
	writeTempFile(t, root, "image.dat", string(binary))

	cfg := demojify.ScanConfig{
		Root:    root,
		Options: demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Fatalf("got %d findings %v, want 1 (text.go only, binary skipped)",
			len(findings), paths)
	}
	if findings[0].Path != "text.go" {
		t.Errorf("finding path = %q, want \"text.go\"", findings[0].Path)
	}
}

func TestScanDirBinaryNulAfterSniffSizeNotSkipped(t *testing.T) {
	root := t.TempDir()

	// Build a text file where the only NUL byte is at position 600
	// (beyond the 512-byte sniff window). It should NOT be treated as binary.
	buf := make([]byte, 700)
	for i := range buf {
		buf[i] = ' ' // fill with spaces (valid text)
	}
	copy(buf, "package p // \U0001F680")
	buf[600] = 0x00
	writeTempFile(t, root, "edge.go", string(buf))

	cfg := demojify.ScanConfig{
		Root:    root,
		Options: demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1 (NUL past sniff window should not skip)",
			len(findings))
	}
}

func TestDefaultScanConfigScansAnyExtension(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "notes.txt", "deploy \U0001F680\n")
	writeTempFile(t, root, "config.ini", "key = \U0001F680\n")
	writeTempFile(t, root, "data.csv", "a,b,\U0001F680\n")
	writeTempFile(t, root, "script.py", "# \U0001F680\n")

	cfg := demojify.DefaultScanConfig()
	cfg.Root = root

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 4 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Errorf("got %d findings %v, want 4 (all file types)", len(findings), paths)
	}
}

func TestScanDir(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, root string) // populate temp dir
		cfg      func(root string) demojify.ScanConfig
		wantLen  int
		wantPath string // first finding path (if wantLen > 0)
	}{
		{
			name: "clean files produce no findings",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "main.go", "package main\n")
				writeTempFile(t, root, "README.md", "# Hello\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen: 0,
		},
		{
			name: "emoji in go file produces finding",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "main.go", "package main // \U0001F680 deploy\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen:  1,
			wantPath: "main.go",
		},
		{
			name: "emoji in markdown produces finding",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "NOTES.md", "# Notes \U0001F4DD\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen:  1,
			wantPath: "NOTES.md",
		},
		{
			name: "test files are exempt by default",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "app_test.go", "package app // \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen: 0,
		},
		{
			name: "skip dirs are excluded",
			setup: func(t *testing.T, root string) {
				vendorDir := writeTempDir(t, root, "vendor")
				writeTempFile(t, vendorDir, "lib.go", "package lib // \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen: 0,
		},
		{
			name: "exempt files are skipped",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "README.md", "# Emoji \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				cfg.ExemptFiles = []string{"README.md"}
				return cfg
			},
			wantLen: 0,
		},
		{
			name: "txt files are scanned by default",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "notes.txt", "deploy \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen:  1,
			wantPath: "notes.txt",
		},
		{
			name: "yaml files are scanned by default",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "config.yaml", "name: test \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen:  1,
			wantPath: "config.yaml",
		},
		{
			name: "json files are scanned by default",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "data.json", "{\"key\": \"\U0001F680\"}\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				return cfg
			},
			wantLen:  1,
			wantPath: "data.json",
		},
		{
			name: "explicit extensions filter excludes non-matching files",
			setup: func(t *testing.T, root string) {
				writeTempFile(t, root, "image.png", "binary \U0001F680\n")
				writeTempFile(t, root, "main.go", "package main\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				cfg := demojify.DefaultScanConfig()
				cfg.Root = root
				cfg.Extensions = []string{".go"} // restrict to .go only
				return cfg
			},
			wantLen: 0, // .png not in explicit extensions list
		},
		{
			name: "nested skip dir is excluded",
			setup: func(t *testing.T, root string) {
				sub := writeTempDir(t, root, filepath.Join("subdir", "vendor"))
				writeTempFile(t, sub, "dep.go", "package dep // \U0001F680\n")
			},
			cfg: func(root string) demojify.ScanConfig {
				return demojify.ScanConfig{
					Root:       root,
					SkipDirs:   []string{"vendor/"},
					Extensions: []string{".go"},
					Options: demojify.Options{
						RemoveEmojis: true,
					},
				}
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)
			cfg := tt.cfg(root)

			findings, err := demojify.ScanDir(cfg)
			if err != nil {
				t.Fatalf("ScanDir: %v", err)
			}
			if len(findings) != tt.wantLen {
				paths := make([]string, len(findings))
				for i, f := range findings {
					paths[i] = f.Path
				}
				t.Fatalf("got %d findings %v, want %d", len(findings), paths, tt.wantLen)
			}
			if tt.wantLen > 0 && findings[0].Path != tt.wantPath {
				t.Errorf("finding path = %q, want %q", findings[0].Path, tt.wantPath)
			}
		})
	}
}

// TestScanDirEmptyRootDefaultsToDot verifies that ScanConfig.Root="" causes
// ScanDir to walk "." (the process working directory) without error.
// The test changes into a small temp directory so the walk is deterministic
// and fast -- it never touches the real repository tree.
func TestScanDirEmptyRootDefaultsToDot(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "clean.go", "package main\n")
	writeTempFile(t, root, "dirty.go", "package main // \U0001F680\n")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cfg := demojify.ScanConfig{
		Root:    "", // should default to "."
		Options: demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir with empty root: %v", err)
	}
	if len(findings) != 1 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Fatalf("got %d findings %v, want 1 (dirty.go)", len(findings), paths)
	}
	if findings[0].Path != "dirty.go" {
		t.Errorf("finding path = %q, want \"dirty.go\"", findings[0].Path)
	}
}

func TestScanDirFindingFields(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "dirty.go", "package dirty // \U0001F680 rocket\n")

	cfg := demojify.DefaultScanConfig()
	cfg.Root = root

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}

	f := findings[0]
	if f.Path != "dirty.go" {
		t.Errorf("Path = %q, want %q", f.Path, "dirty.go")
	}
	if !f.HasEmoji {
		t.Error("HasEmoji = false, want true")
	}
	if f.Original != "package dirty // \U0001F680 rocket\n" {
		t.Errorf("Original = %q, unexpected", f.Original)
	}
	if f.Cleaned == f.Original {
		t.Error("Cleaned should differ from Original")
	}
}

func TestScanDirErrorOnBadRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir", "deep")
	cfg := demojify.ScanConfig{
		Root: missing,
	}
	_, err := demojify.ScanDir(cfg)
	if err == nil {
		t.Error("expected error for nonexistent root, got nil")
	}
}

func TestScanDirUnreadableFile(t *testing.T) {
	if isWindows() {
		t.Skip("os.Chmod cannot remove read permission on Windows")
	}

	root := t.TempDir()
	path := writeTempFile(t, root, "secret.md", "\u2705 hidden\n")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}

	// When running as root (or with equivalent privileges), chmod 000 may
	// not actually prevent reads. Detect this and skip.
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("chmod 000 did not prevent reading (likely running as root)")
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = root
	cfg.ExemptSuffixes = nil

	_, err := demojify.ScanDir(cfg)
	if err == nil {
		t.Error("expected error for unreadable file, got nil")
	}
}

// TestScanDirSymlink verifies that ScanDir follows symbolic links to
// regular files and reports findings in them.
func TestScanDirSymlink(t *testing.T) {
	if isWindows() {
		t.Skip("creating symlinks on Windows requires elevated privileges")
	}

	root := t.TempDir()
	sub := writeTempDir(t, root, "real")
	writeTempFile(t, sub, "status.md", "\u2705 passed\n")

	// Create a symlink in root pointing to the real file.
	link := filepath.Join(root, "link.md")
	if err := os.Symlink(filepath.Join(sub, "status.md"), link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = root
	cfg.ExemptSuffixes = nil

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	// Should find the symlinked file (WalkDir follows symlinks to files).
	found := false
	for _, f := range findings {
		if strings.Contains(f.Path, "link.md") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ScanDir to follow symlink and report finding for link.md")
	}
}

func TestScanDirContext(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "emoji.md", "Hello \U0001F680 World\n")
	writeTempFile(t, root, "clean.md", "Hello World\n")

	t.Run("normal context returns findings", func(t *testing.T) {
		cfg := demojify.DefaultScanConfig()
		cfg.Root = root
		cfg.Extensions = []string{".md"}

		findings, err := demojify.ScanDirContext(context.Background(), cfg)
		if err != nil {
			t.Fatalf("ScanDirContext: %v", err)
		}
		if len(findings) != 1 {
			t.Errorf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("canceled context returns error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		cfg := demojify.DefaultScanConfig()
		cfg.Root = root
		cfg.Extensions = []string{".md"}

		_, err := demojify.ScanDirContext(ctx, cfg)
		if err == nil {
			t.Fatal("expected error from canceled context, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("expected context canceled error, got: %v", err)
		}
	})
}

// TestScanDirEmptyFile verifies that ScanDir handles a directory containing
// an empty file and a non-empty, emoji-free file without reporting findings.
func TestScanDirEmptyFile(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "empty.go", "")
	writeTempFile(t, root, "notempty.go", "package main\n")

	cfg := demojify.DefaultScanConfig()
	cfg.Root = root

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 0 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Errorf("got %d findings %v, want 0 (neither file has emoji)", len(findings), paths)
	}
}

// TestScanDirMaxFileBytesExactBoundary verifies the edge case where a file
// is exactly at the MaxFileBytes limit. A file at the boundary should be
// scanned (not skipped).

func TestScanDirMaxFileBytesExactBoundary(t *testing.T) {
	root := t.TempDir()

	// Create a file with emoji whose total size matches the limit exactly.
	content := "emoji \U0001F680 here\n" // 16 bytes
	writeTempFile(t, root, "exact.txt", content)

	cfg := demojify.ScanConfig{
		Root:         root,
		MaxFileBytes: int64(len(content)), // exactly at the limit
		Options:      demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1 (file at exact boundary should be scanned)", len(findings))
	}

	// File one byte over the limit should be skipped.
	cfg.MaxFileBytes = int64(len(content)) - 1
	findings, err = demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0 (file exceeds limit by 1 byte)", len(findings))
	}
}

// TestScanDirConcurrent verifies that ScanDir is safe for concurrent use
// when called from multiple goroutines on separate directory trees.
func TestScanDirConcurrent(t *testing.T) {
	const goroutines = 10
	roots := make([]string, goroutines)
	for i := 0; i < goroutines; i++ {
		roots[i] = t.TempDir()
		writeTempFile(t, roots[i], "file.md", "Hello \U0001F680 World\n")
		writeTempFile(t, roots[i], "clean.md", "Hello World\n")
	}

	errCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			cfg := demojify.DefaultScanConfig()
			cfg.Root = roots[idx]
			cfg.Extensions = []string{".md"}
			findings, err := demojify.ScanDir(cfg)
			if err != nil {
				errCh <- err
				return
			}
			if len(findings) != 1 {
				errCh <- fmt.Errorf("expected 1 finding, got %d", len(findings))
				return
			}
			errCh <- nil
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("goroutine %d: ScanDir error: %v", i, err)
		}
	}
}

// TestScanDirPreservesCRLF is a regression test for the bug where emoji removal
// silently converted CRLF (\r\n) line endings to LF (\n) even when
// NormalizeWhitespace was false. After the fix, CRLF files with emoji must have
// their line endings preserved in the Cleaned output; only the emoji is removed.
func TestScanDirPreservesCRLF(t *testing.T) {
	root := t.TempDir()
	// File with CRLF line endings and an emoji on the second line.
	const content = "line one\r\n" +
		"deploy \U0001F680 done\r\n" +
		"line three\r\n"
	writeTempFile(t, root, "deploy.md", content)

	cfg := demojify.ScanConfig{
		Root:    root,
		Options: demojify.Options{RemoveEmojis: true},
		// NormalizeWhitespace deliberately left false (the default).
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}

	cleaned := findings[0].Cleaned

	// The emoji must be gone.
	if strings.Contains(cleaned, "\U0001F680") {
		t.Errorf("emoji still present in cleaned output: %q", cleaned)
	}

	// CRLF line endings must be preserved.
	if !strings.Contains(cleaned, "\r\n") {
		t.Errorf("CRLF line endings were lost (converted to LF): %q", cleaned)
	}

	// No bare LF should remain (every newline must be \r\n).
	cleanedNoCRLF := strings.ReplaceAll(cleaned, "\r\n", "")
	if strings.Contains(cleanedNoCRLF, "\n") {
		t.Errorf("cleaned output contains bare LF line endings: %q", cleaned)
	}

	// The surrounding text must be intact.
	if !strings.Contains(cleaned, "line one") || !strings.Contains(cleaned, "deploy") || !strings.Contains(cleaned, "done") {
		t.Errorf("surrounding text was altered in cleaned output: %q", cleaned)
	}
}

// TestScanDirPreservesCRLFCleanFile verifies that a CRLF file with no emoji
// is not reported as a finding (i.e., CRLF is not altered for clean files).
func TestScanDirPreservesCRLFCleanFile(t *testing.T) {
	root := t.TempDir()
	const content = "no emoji here\r\nall clean\r\n"
	writeTempFile(t, root, "clean.md", content)

	cfg := demojify.ScanConfig{
		Root:    root,
		Options: demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings for clean CRLF file, want 0", len(findings))
	}
}
