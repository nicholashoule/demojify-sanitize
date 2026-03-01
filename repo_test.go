package demojify_test

// repo_test.go enforces repository hygiene using the module's own public API.
// These tests walk the real working directory and are therefore non-hermetic.
//
// Go modules exclude *_test.go files when consumed as a dependency (via
// go get or go mod vendor), so these tests only run for developers who
// have cloned this repository -- never for downstream consumers.
//
// Design intent:
//
//   *_test.go files are INTENTIONAL emoji sources. They contain literal emoji
//   as input data that exercises the module's detection and removal logic.
//   These files are EXEMPT from hygiene enforcement.
//
//   Non-test Go source files and all Markdown documentation MUST be
//   emoji-free. If an AI agent writes emoji into any of these files, the
//   tests below will catch it and identify the file and the fix.
//
// Enforcement dogfoods [ScanDir] and [ScanFile] rather than reimplementing
// directory walking, proving the scanner API works on a real repository.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// TestRepoProductionSourceFilesEmojiClean uses [ScanDir] to assert that every
// non-test Go file contains no literal emoji. Production source must be
// emoji-free; test files are deliberately exempted by [ScanConfig.ExemptSuffixes].
//
// If this test fails, an AI agent has introduced emoji into production source.
// Fix: run demojify.SanitizeFile on the reported path.
func TestRepoProductionSourceFilesEmojiClean(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	cfg.Root = "."
	cfg.Extensions = []string{".go"}
	// DefaultScanConfig already exempts _test.go via ExemptSuffixes.

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	for _, f := range findings {
		t.Errorf("%s contains emoji (detected by ScanDir)\n"+
			"Fix: run demojify.SanitizeFile(%q, demojify.Options{RemoveEmojis: true})", f.Path, f.Path)
	}
}

// TestRepoAllDocsEmojiClean uses [ScanDir] to assert that every Markdown file
// (excluding docs/) contains no literal emoji. Covers README.md, CHANGELOG.md,
// CONTRIBUTING.md, SECURITY.md, .github/ files, and any nested Markdown.
//
// If this test fails, apply demojify.SanitizeFile to the reported file.
func TestRepoAllDocsEmojiClean(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	cfg.Root = "."
	cfg.Extensions = []string{".md"}
	cfg.SkipDirs = append(cfg.SkipDirs, "docs/")

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	for _, f := range findings {
		t.Errorf("%s contains emoji (detected by ScanDir)\n"+
			"Fix: run demojify.SanitizeFile(%q, demojify.Options{RemoveEmojis: true})", f.Path, f.Path)
	}
}

// TestRepoProductionFilesIdempotent uses [ScanDir] to assert that running
// Sanitize (emoji removal only) on every non-test Go source file and every
// Markdown file is a no-op -- the files are already clean.
//
// Whitespace normalization is disabled because gofmt owns Go formatting.
//
// If this test fails, Sanitize changes the file -- it contains emoji.
// Write the Sanitize output back to fix it.
func TestRepoProductionFilesIdempotent(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	cfg.Root = "."
	cfg.Extensions = []string{".go", ".md"}
	cfg.SkipDirs = append(cfg.SkipDirs, "docs/")
	// RemoveEmojis only; NormalizeWhitespace stays false (default).

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	for _, f := range findings {
		t.Errorf("%s: Sanitize modifies this file -- it contains emoji.\n"+
			"Fix: run demojify.SanitizeFile(%q, demojify.Options{RemoveEmojis: true})", f.Path, f.Path)
	}
}

// TestRepoTestFilesContainEmoji is a meta-test that verifies at least one
// *_test.go file contains literal emoji. This confirms that exemptions are
// load-bearing: test files ARE the module's input data, and they must contain
// real emoji to exercise detection and removal.
func TestRepoTestFilesContainEmoji(t *testing.T) {
	found := false
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Logf("skip %s: %v", path, readErr)
			return nil
		}
		if demojify.ContainsEmoji(string(data)) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if !found {
		t.Error("no test file contains literal emoji -- test data is missing;\n" +
			"unit test files must contain real emoji codepoints to prove the module processes them")
	}
}
