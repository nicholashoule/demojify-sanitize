package demojify_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// TestScanDirNormalizeUnconditional verifies that NormalizeWhitespace
// runs unconditionally when enabled, producing findings for any file
// whose whitespace can be collapsed -- regardless of whether the
// emoji/replacement step also changed the content.
func TestScanDirNormalizeUnconditional(t *testing.T) {
	t.Run("whitespace-only file is a finding when emoji removal and normalization are active", func(t *testing.T) {
		root := t.TempDir()
		// Irregular whitespace (double spaces) but no emoji.
		writeTempFile(t, root, "spaces.txt", "hello  world\n")

		cfg := demojify.ScanConfig{
			Root: root,
			Options: demojify.Options{
				RemoveEmojis:        true,
				NormalizeWhitespace: true,
			},
		}
		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1 (normalization runs unconditionally)",
				len(findings))
		}
		want := "hello world"
		if findings[0].Cleaned != want {
			t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
		}
	})

	t.Run("file with emoji and irregular whitespace is normalized", func(t *testing.T) {
		root := t.TempDir()
		// Emoji + double spaces: both should be cleaned.
		writeTempFile(t, root, "both.txt", "\U0001F680  deployed  ok\n")

		cfg := demojify.ScanConfig{
			Root: root,
			Options: demojify.Options{
				RemoveEmojis:        true,
				NormalizeWhitespace: true,
			},
		}
		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1", len(findings))
		}
		// Emoji removed, then double spaces collapsed.
		want := "deployed ok"
		if findings[0].Cleaned != want {
			t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
		}
	})

	t.Run("standalone normalization (no emoji removal) applies unconditionally", func(t *testing.T) {
		root := t.TempDir()
		writeTempFile(t, root, "spaces.txt", "hello  world\n")

		cfg := demojify.ScanConfig{
			Root: root,
			Options: demojify.Options{
				RemoveEmojis:        false,
				NormalizeWhitespace: true,
			},
		}
		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1 (standalone normalization)", len(findings))
		}
		want := "hello world"
		if findings[0].Cleaned != want {
			t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
		}
	})

	t.Run("replacement path: whitespace-only file is a finding", func(t *testing.T) {
		root := t.TempDir()
		writeTempFile(t, root, "spaces.txt", "hello  world\n")

		cfg := demojify.ScanConfig{
			Root:         root,
			Replacements: map[string]string{"\u2705": "[PASS]"},
			Options: demojify.Options{
				NormalizeWhitespace: true,
			},
		}
		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1 (normalization runs unconditionally)",
				len(findings))
		}
		want := "hello world"
		if findings[0].Cleaned != want {
			t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
		}
	})

	t.Run("replacement path: file with emoji and whitespace is normalized", func(t *testing.T) {
		root := t.TempDir()
		writeTempFile(t, root, "status.txt", "\u2705  passed  ok\n")

		cfg := demojify.ScanConfig{
			Root:         root,
			Replacements: map[string]string{"\u2705": "[PASS]"},
			Options: demojify.Options{
				NormalizeWhitespace: true,
			},
		}
		findings, err := demojify.ScanDir(cfg)
		if err != nil {
			t.Fatalf("ScanDir: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("got %d findings, want 1", len(findings))
		}
		want := "[PASS] passed ok"
		if findings[0].Cleaned != want {
			t.Errorf("Cleaned = %q, want %q", findings[0].Cleaned, want)
		}
	})
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
	want := "[PASS] done launch\n"
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
	if m0.Sequence != "\u2705" {
		t.Errorf("Matches[0].Sequence = %q, want checkmark", m0.Sequence)
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
	if m1.Sequence != "\U0001F680" {
		t.Errorf("Matches[1].Sequence = %q, want rocket", m1.Sequence)
	}
	if m1.Replacement != "" {
		t.Errorf("Matches[1].Replacement = %q, want empty (unmapped)", m1.Replacement)
	}
	if m1.Line != 2 {
		t.Errorf("Matches[1].Line = %d, want 2", m1.Line)
	}
}

// TestScanDirCollectMatchesVariationSelector verifies that buildMatches
// attributes a variation-selector sequence (e.g., U+26A0 U+FE0F) to the
// combined key in the replacements map rather than to the bare codepoint.
func TestScanDirCollectMatchesVariationSelector(t *testing.T) {
	root := t.TempDir()
	// U+26A0 U+FE0F: warning sign + variation selector 16
	writeTempFile(t, root, "warn.txt", "\u26a0\ufe0f critical issue")

	repl := map[string]string{
		"\u26a0\ufe0f": "WARNING:", // combined key
		"\u26a0":       "WARN:",    // bare codepoint fallback
	}
	cfg := demojify.ScanConfig{
		Root:           root,
		CollectMatches: true,
		Replacements:   repl,
		Options:        demojify.Options{RemoveEmojis: true},
	}
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	if len(findings[0].Matches) != 1 {
		t.Fatalf("Matches count = %d, want 1 (combined key)", len(findings[0].Matches))
	}
	m := findings[0].Matches[0]
	if m.Sequence != "\u26a0\ufe0f" {
		t.Errorf("Sequence = %q, want combined sequence", m.Sequence)
	}
	if m.Replacement != "WARNING:" {
		t.Errorf("Replacement = %q, want WARNING:", m.Replacement)
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

// TestScanDirCollectMatchesNonEmojiKey verifies that replacement keys outside
// the emoji regex (such as arrows U+2192) are still recorded as Matches, and
// that the resulting Finding has HasEmoji == false when no actual emoji is
// present in the file.
func TestScanDirCollectMatchesNonEmojiKey(t *testing.T) {
	root := t.TempDir()
	// U+2192 (RIGHTWARDS ARROW) is not in emojiRE, but is a valid
	// replacement key in DefaultReplacements.
	writeTempFile(t, root, "arrows.txt", "step \u2192 next")

	repl := map[string]string{"\u2192": "->"}
	cfg := demojify.ScanConfig{
		Root:           root,
		CollectMatches: true,
		Replacements:   repl,
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
	// Arrow is not an emoji per emojiRE, so HasEmoji must be false.
	if f.HasEmoji {
		t.Error("HasEmoji = true, want false (arrow is not in emojiRE)")
	}
	if len(f.Matches) != 1 {
		t.Fatalf("Matches count = %d, want 1", len(f.Matches))
	}
	m := f.Matches[0]
	if m.Sequence != "\u2192" {
		t.Errorf("Sequence = %q, want U+2192 arrow", m.Sequence)
	}
	if m.Replacement != "->" {
		t.Errorf("Replacement = %q, want \"->\"", m.Replacement)
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

func TestFindMatchesInFile(t *testing.T) {
	repl := demojify.DefaultReplacements()

	t.Run("file with emoji returns matches with line and column", func(t *testing.T) {
		dir := t.TempDir()
		// Line 1: checkmark at column 0; line 2: cross mark at column 0
		path := writeTempFile(t, dir, "doc.md", "\u2705 passed\n\u274c failed\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if len(matches) != 2 {
			t.Fatalf("got %d matches, want 2", len(matches))
		}

		m0 := matches[0]
		if m0.Sequence != "\u2705" {
			t.Errorf("matches[0].Sequence = %q, want checkmark", m0.Sequence)
		}
		if m0.Replacement != "[PASS]" {
			t.Errorf("matches[0].Replacement = %q, want [PASS]", m0.Replacement)
		}
		if m0.Line != 1 {
			t.Errorf("matches[0].Line = %d, want 1", m0.Line)
		}
		if m0.Column != 0 {
			t.Errorf("matches[0].Column = %d, want 0", m0.Column)
		}
		if m0.Context == "" {
			t.Error("matches[0].Context should not be empty")
		}

		m1 := matches[1]
		if m1.Line != 2 {
			t.Errorf("matches[1].Line = %d, want 2", m1.Line)
		}
		if m1.Sequence != "\u274c" {
			t.Errorf("matches[1].Sequence = %q, want cross mark", m1.Sequence)
		}
		if m1.Replacement != "[FAIL]" {
			t.Errorf("matches[1].Replacement = %q, want [FAIL]", m1.Replacement)
		}
	})

	t.Run("file with no emoji returns nil", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "clean.txt", "This file has no emoji\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if matches != nil {
			t.Errorf("got %d matches, want nil for clean file", len(matches))
		}
	})

	t.Run("unmapped emoji has empty replacement", func(t *testing.T) {
		dir := t.TempDir()
		// U+1F600 GRINNING FACE -- not in DefaultReplacements, so replacement is empty.
		path := writeTempFile(t, dir, "log.txt", "\U0001F600 hello\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("expected at least one match for grinning face emoji")
		}
		// Grinning face is not in DefaultReplacements; replacement should be empty.
		if matches[0].Replacement != "" {
			t.Errorf("Replacement = %q, want empty for unmapped emoji", matches[0].Replacement)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir", "no-file.txt")
		_, err := demojify.FindMatchesInFile(missing, repl)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("binary file is skipped", func(t *testing.T) {
		dir := t.TempDir()
		// Embed a NUL byte in the first 512 bytes so isBinary returns true.
		path := writeTempFile(t, dir, "data.bin", "hello\x00\u2705 world\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if matches != nil {
			t.Errorf("got %d matches, want nil for binary file", len(matches))
		}
	})
}

// TestScanDirUnreadableFile verifies that ScanDir returns an error when it
// encounters a file that cannot be read. On Windows, os.Chmod does not
// support removing read permission, so this test is skipped there.
