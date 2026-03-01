package demojify_test

import (
	"os"
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// writeTempFile creates a file inside dir with the given name and content.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// writeTempDir creates a subdirectory inside dir.
func writeTempDir(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	return path
}

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
	_, err := demojify.ScanFile("/nonexistent/path/file.go", demojify.DefaultOptions())
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestScanDirErrorOnBadRoot(t *testing.T) {
	cfg := demojify.ScanConfig{
		Root: "/nonexistent/directory/that/should/not/exist",
	}
	_, err := demojify.ScanDir(cfg)
	if err == nil {
		t.Error("expected error for nonexistent root, got nil")
	}
}
