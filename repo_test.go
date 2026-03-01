package demojify_test

// repo_test.go proves this module's value against its own repository.
//
// Design intent:
//
//   *_test.go files are INTENTIONAL emoji sources. They contain literal emoji
//   as input data that exercises the module's detection and removal logic.
//   These files are EXEMPT from hygiene enforcement -- they are the proof that
//   the module works on real-world emoji codepoints.
//
//   Non-test Go source files and all Markdown documentation (except README.md)
//   MUST be emoji-free. If an AI agent writes emoji into any of these files
//   while ignoring emoji-prevention.md or copilot-instructions.md, the tests
//   below will catch it and identify the file and the fix.
//
// Enforcement uses only this module's own API -- ContainsEmoji to detect,
// Sanitize to remediate -- demonstrating the same operations callers apply.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// skipDirs are never walked during repo hygiene checks.
// Add any generated or third-party directories here.
var skipDirs = []string{".git/", "vendor/", "node_modules/"}

// exemptMarkdown lists Markdown files that intentionally contain literal emoji
// in code-fence examples to illustrate library behaviour.
var exemptMarkdown = map[string]bool{
	"README.md": true,
}

// readRepoFiles walks from root and returns paths to all files with the given
// extension, excluding directories listed in skipDirs.
func readRepoFiles(t *testing.T, root, ext string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		norm := filepath.ToSlash(path)
		for _, skip := range skipDirs {
			if strings.HasPrefix(norm, skip) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ext) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return paths
}

// isTestFile reports whether path is a Go test file (*_test.go).
// Test files contain intentional emoji as module input data and are exempt
// from hygiene enforcement.
func isTestFile(path string) bool {
	return strings.HasSuffix(filepath.ToSlash(path), "_test.go")
}

// TestRepoProductionSourceFilesEmojiClean asserts that every non-test Go file
// contains no literal emoji. Production source must be emoji-free; emoji in
// test files is deliberate and is excluded from this check.
//
// If this test fails, an AI agent has introduced emoji into production source.
// Fix: read the file, apply demojify.Sanitize, and write back.
func TestRepoProductionSourceFilesEmojiClean(t *testing.T) {
	for _, path := range readRepoFiles(t, ".", ".go") {
		if isTestFile(path) {
			continue // test files contain intentional emoji test data
		}
		path := path
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if demojify.ContainsEmoji(string(data)) {
				clean := demojify.Sanitize(string(data), demojify.DefaultOptions())
				t.Errorf("%s contains emoji (detected by ContainsEmoji)\n"+
					"Run demojify.Sanitize on the file to remove it.\n"+
					"Sanitize would produce:\n%s", path, clean)
			}
		})
	}
}

// TestRepoAllDocsEmojiClean asserts that every Markdown file (except exempted
// ones) contains no literal emoji. Covers docs/, .github/instructions/,
// .github/ISSUE_TEMPLATE/, and all .github/ root files -- every file an AI
// agent might write to when ignoring emoji-prevention.md.
//
// If this test fails, apply demojify.Sanitize to the reported file.
func TestRepoAllDocsEmojiClean(t *testing.T) {
	for _, path := range readRepoFiles(t, ".", ".md") {
		if exemptMarkdown[filepath.Base(path)] {
			continue
		}
		path := path
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if demojify.ContainsEmoji(string(data)) {
				clean := demojify.Sanitize(string(data), demojify.DefaultOptions())
				t.Errorf("%s contains emoji (detected by ContainsEmoji)\n"+
					"Run demojify.Sanitize on the file to remove it.\n"+
					"Sanitize would produce:\n%s", path, clean)
			}
		})
	}
}

// TestRepoProductionFilesIdempotent asserts that running Sanitize (emoji removal
// + AI-clutter removal) on every non-test Go source file and every non-exempted
// Markdown file is a no-op -- the files are already clean.
//
// Test files are exempt: their contents ARE the module's input data.
// Whitespace normalization is excluded: gofmt owns Go source formatting.
//
// If this test fails, Sanitize changes the file -- it contains emoji or AI
// preamble clutter. Write the Sanitize output back to the file to fix it.
func TestRepoProductionFilesIdempotent(t *testing.T) {
	opts := demojify.Options{
		RemoveEmojis:    true,
		RemoveAIClutter: true,
	}

	check := func(path string) {
		t.Helper()
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			original := string(data)
			cleaned := demojify.Sanitize(original, opts)
			if cleaned != original {
				t.Errorf("%s: Sanitize modifies this file -- it contains emoji or AI preamble clutter.\n"+
					"Fix: write demojify.Sanitize(content, demojify.DefaultOptions()) back to the file.", path)
			}
		})
	}

	for _, path := range readRepoFiles(t, ".", ".go") {
		if !isTestFile(path) {
			check(path)
		}
	}
	for _, path := range readRepoFiles(t, ".", ".md") {
		if !exemptMarkdown[filepath.Base(path)] {
			check(path)
		}
	}
}

// TestRepoTestFilesContainEmoji is a meta-test that verifies test files DO
// contain literal emoji. This confirms the exemption above is load-bearing:
// test files are the module's input data, and they must contain real emoji
// to exercise the detection and removal logic.
func TestRepoTestFilesContainEmoji(t *testing.T) {
	found := false
	for _, path := range readRepoFiles(t, ".", ".go") {
		if !isTestFile(path) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if demojify.ContainsEmoji(string(data)) {
			found = true
			break
		}
	}
	if !found {
		t.Error("no test file contains literal emoji -- test data is missing;\n" +
			"unit test files must contain real emoji codepoints to prove the module processes them")
	}
}

// TestRepoAgentOutputRemediation proves that the module detects and fully
// remediates the kind of content a rogue AI agent produces when it ignores
// emoji-prevention.md or copilot-instructions.md. This is the core value
// demonstration: ContainsEmoji catches the violation, Sanitize removes emoji
// and AI preamble in one call, and the result is idempotent.
func TestRepoAgentOutputRemediation(t *testing.T) {
	// Simulate text a rogue agent writes, mixing preamble, emoji, and real content.
	rogueOutput := strings.Join([]string{
		"Certainly! Here is the updated documentation.",
		"",
		"\U0001F680 Deployment",
		"",
		"Run the following command to deploy:",
		"",
		"    go build ./...",
		"",
		"I hope this helps! \U0001F4CA",
		"",
		"Feel free to ask if you need further assistance.",
	}, "\n")

	// Step 1: ContainsEmoji detects the violation before the file is written.
	if !demojify.ContainsEmoji(rogueOutput) {
		t.Fatal("ContainsEmoji: expected true for rogue agent output, got false")
	}

	// Step 2: Sanitize remediates emoji, preamble, and trailing clutter in one call.
	clean := demojify.Sanitize(rogueOutput, demojify.DefaultOptions())

	// Step 3: output is now emoji-free.
	if demojify.ContainsEmoji(clean) {
		t.Errorf("after Sanitize, output still contains emoji:\n%s", clean)
	}

	// Step 4: AI preamble and filler phrases are removed.
	for _, banned := range []string{"Certainly!", "I hope this helps", "Feel free to ask"} {
		if strings.Contains(clean, banned) {
			t.Errorf("after Sanitize, output still contains banned phrase %q:\n%s", banned, clean)
		}
	}

	// Step 5: substantive content is preserved.
	for _, required := range []string{"Deployment", "go build ./..."} {
		if !strings.Contains(clean, required) {
			t.Errorf("after Sanitize, output is missing expected content %q:\n%s", required, clean)
		}
	}

	// Step 6: running Sanitize again produces identical output -- idempotent.
	if twice := demojify.Sanitize(clean, demojify.DefaultOptions()); twice != clean {
		t.Errorf("Sanitize is not idempotent:\nfirst:  %q\nsecond: %q", clean, twice)
	}
}
