package demojify_test

import (
	"os"
	"path/filepath"
	"strings"
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
	copy(large, []byte("\U0001F680\n"))
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

func TestScanDirReplacements(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "status.txt", "build \u2705 passed\n")

	cfg := demojify.ScanConfig{
		Root:         root,
		Replacements: map[string]string{"\u2705": "[PASS]"},
		Options:      demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	want := "build [PASS] passed\n"
	if findings[0].Cleaned != want {
		t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
	}
}

func TestScanDirReplacementsUnmappedEmojiStripped(t *testing.T) {
	root := t.TempDir()
	// checkmark is mapped; rocket is not -- should be stripped
	writeTempFile(t, root, "out.txt", "\u2705 done \U0001F680 launch\n")

	cfg := demojify.ScanConfig{
		Root:         root,
		Replacements: map[string]string{"\u2705": "[PASS]"},
		Options:      demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	want := "[PASS] done  launch\n"
	if findings[0].Cleaned != want {
		t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
	}
}

func TestScanDirCollectMatches(t *testing.T) {
	root := t.TempDir()
	// Line 1: two emoji -- checkmark (mapped) and rocket (unmapped)
	writeTempFile(t, root, "log.txt", "\u2705 pass\n\U0001F680 launch\n")

	cfg := demojify.ScanConfig{
		Root:           root,
		CollectMatches: true,
		Replacements:   map[string]string{"\u2705": "[PASS]"},
		Options:        demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	f := findings[0]
	if len(f.Matches) != 2 {
		t.Fatalf("Matches count = %d, want 2", len(f.Matches))
	}

	// First match: checkmark on line 1
	m0 := f.Matches[0]
	if m0.Emoji != "\u2705" {
		t.Errorf("Matches[0].Emoji = %q, want checkmark", m0.Emoji)
	}
	if m0.Replacement != "[PASS]" {
		t.Errorf("Matches[0].Replacement = %q, want [PASS]", m0.Replacement)
	}
	if m0.Line != 1 {
		t.Errorf("Matches[0].Line = %d, want 1", m0.Line)
	}
	if m0.Column != 0 {
		t.Errorf("Matches[0].Column = %d, want 0", m0.Column)
	}
	if m0.Context != "\u2705 pass" {
		t.Errorf("Matches[0].Context = %q, want checkmark line", m0.Context)
	}

	// Second match: rocket on line 2
	m1 := f.Matches[1]
	if m1.Emoji != "\U0001F680" {
		t.Errorf("Matches[1].Emoji = %q, want rocket", m1.Emoji)
	}
	if m1.Replacement != "" {
		t.Errorf("Matches[1].Replacement = %q, want empty (unmapped)", m1.Replacement)
	}
	if m1.Line != 2 {
		t.Errorf("Matches[1].Line = %d, want 2", m1.Line)
	}
}

func TestScanDirCollectMatchesFalseGivesNilMatches(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "a.txt", "\u2705 done\n")

	cfg := demojify.ScanConfig{
		Root:           root,
		CollectMatches: false,
		Options:        demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	if findings[0].Matches != nil {
		t.Errorf("Matches = %v, want nil when CollectMatches is false", findings[0].Matches)
	}
}

// TestScanDirFilterByExtension verifies that ScanDir with an Extensions filter
// returns only files matching those extensions, mirroring the
// directory-scanning integration pattern used by emoji-cleaner tooling.
func TestScanDirFilterByExtension(t *testing.T) {
	root := t.TempDir()
	subDir := writeTempDir(t, root, "sub")

	writeTempFile(t, root, "file1.md", "\u2705 Markdown file\n")
	writeTempFile(t, root, "file2.txt", "\u274c Text file\n")
	writeTempFile(t, root, "file3.go", "// No emojis here\n")
	writeTempFile(t, subDir, "file4.md", "\u26a0 Nested file\n")

	// All emoji-containing files detected when no extension filter.
	cfg := demojify.DefaultScanConfig()
	cfg.Root = root
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir (no filter): %v", err)
	}
	if len(findings) != 3 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Errorf("ScanDir (no filter): got %d findings %v, want 3", len(findings), paths)
	}

	// Restrict to .md only -- should find file1.md and sub/file4.md.
	cfg.Extensions = []string{".md"}
	findings, err = demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir (.md filter): %v", err)
	}
	if len(findings) != 2 {
		paths := make([]string, len(findings))
		for i, f := range findings {
			paths[i] = f.Path
		}
		t.Errorf("ScanDir (.md filter): got %d findings %v, want 2", len(findings), paths)
	}
	for _, f := range findings {
		if !strings.HasSuffix(f.Path, ".md") {
			t.Errorf("non-.md file in findings: %s", f.Path)
		}
	}
}

// TestScanDirReplaceAndSave verifies the ScanDir + ReplaceFile integration
// pattern: scan a directory, apply substitutions to each dirty file, confirm
// total replacement count and final file content.
func TestScanDirReplaceAndSave(t *testing.T) {
	root := t.TempDir()
	subDir := writeTempDir(t, root, "sub")

	writeTempFile(t, root, "file1.md", "\u2705 Success\n")
	writeTempFile(t, root, "file2.txt", "\u274c Failure\n")
	writeTempFile(t, root, "file3.md", "No emojis\n")
	writeTempFile(t, subDir, "file4.md", "\u26a0 Warning \u2705 Done\n")

	repl := demojify.DefaultReplacements()

	// Scan .md files only -- file3.md is clean so not a finding.
	cfg := demojify.DefaultScanConfig()
	cfg.Root = root
	cfg.Extensions = []string{".md"}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("ScanDir: got %d findings, want 2", len(findings))
	}

	totalCount := 0
	for _, f := range findings {
		absPath := filepath.Join(root, filepath.FromSlash(f.Path))
		count, err := demojify.ReplaceFile(absPath, repl)
		if err != nil {
			t.Fatalf("ReplaceFile(%s): %v", f.Path, err)
		}
		totalCount += count
	}

	// file1.md: 1 substitution; sub/file4.md: 2 substitutions.
	if totalCount != 3 {
		t.Errorf("total replacements = %d, want 3", totalCount)
	}

	data, err := os.ReadFile(filepath.Join(root, "file1.md"))
	if err != nil {
		t.Fatalf("ReadFile file1.md: %v", err)
	}
	if string(data) != "[PASS] Success\n" {
		t.Errorf("file1.md content = %q, want \"[PASS] Success\\n\"", string(data))
	}
}
