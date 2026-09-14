// Repository hygiene tests that run the scanner against this checkout.
//
// These tests are non-hermetic: they walk the real working directory. Go
// excludes *_test.go files from downstream consumers, so they run only for
// developers who cloned the repository.
//
// Policy enforced:
//   - Non-test Go source and all Markdown must be emoji-free, and Sanitize
//     must be a no-op on them.
//   - *_test.go files are exempt; they are the intended source of literal
//     emoji test input, and one test asserts that at least one contains some.
//   - scripts/hooks/pre-commit must filter by extension and run the CLI
//     from this working tree.
//
// Enforcement dogfoods ScanDir rather than reimplementing the walk, proving
// the scanner works on a real repository.

package demojify_test

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
// (including docs/) contains no literal emoji. Covers README.md, CHANGELOG.md,
// CONTRIBUTING.md, SECURITY.md, .github/ files, and all docs/ Markdown.
//
// If this test fails, apply demojify.SanitizeFile to the reported file.
// To include illustrative emoji in a doc, use Unicode escape sequences
// (e.g. \U0001F680) so ContainsEmoji reports false on the source file.
func TestRepoAllDocsEmojiClean(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	cfg.Root = "."
	cfg.Extensions = []string{".md"}
	// docs/ is intentionally included: all production docs must be emoji-free.
	// Use Unicode escape sequences (\U0001F680) when writing about emoji in docs.
	// tmp/ directories hold intentionally emoji-laden test fixtures; skip them.
	cfg.SkipDirs = append(cfg.SkipDirs, "tmp/")

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
	// docs/ is included: production docs must be clean and idempotent.
	// RemoveEmojis only; NormalizeWhitespace stays false (default).
	// tmp/ directories hold intentionally emoji-laden test fixtures; skip them.
	cfg.SkipDirs = append(cfg.SkipDirs, "tmp/")

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	for _, f := range findings {
		t.Errorf("%s: Sanitize modifies this file -- it contains emoji.\n"+
			"Fix: run demojify.SanitizeFile(%q, demojify.Options{RemoveEmojis: true})", f.Path, f.Path)
	}
}

// TestHookDemojifyUsesExtensionFilter ensures the shipped pre-commit hook limits
// demojify scanning to known text file types so compressed binary assets do not
// produce false-positive emoji matches.
func TestHookDemojifyUsesExtensionFilter(t *testing.T) {
	data, err := os.ReadFile("scripts/hooks/pre-commit")
	if err != nil {
		t.Fatalf("read pre-commit hook: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "demojify_exts=") || !strings.Contains(s, "demojify_filters=") {
		t.Fatal("scripts/hooks/pre-commit must define demojify_exts and demojify_filters for filtered scanning")
	}
	if !strings.Contains(s, "-exts \"$demojify_exts\"") {
		t.Fatal("scripts/hooks/pre-commit must pass -exts \"$demojify_exts\" to demojify")
	}
	if !strings.Contains(s, "$demojify_filters") {
		t.Fatal("scripts/hooks/pre-commit must pass $demojify_filters to demojify")
	}
	// The hook runs the CLI from this working tree rather than a published
	// release: this repository is the CLI's source, so the hook dogfoods the
	// current code and never lags a release. Downstream hooks pin a tag; see
	// docs/git-hooks.md.
	if !strings.Contains(s, `demojify_ref="./cmd/demojify"`) {
		t.Fatal(`scripts/hooks/pre-commit must set demojify_ref="./cmd/demojify" to run the CLI from the working tree`)
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
