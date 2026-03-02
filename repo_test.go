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
	"bufio"
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

// TestRepoLicenseStartsOnLineOne verifies that LICENSE begins with non-whitespace
// on its very first line. pkg.go.dev's licensecheck scanner reads from byte 0;
// a leading blank line prevents Apache-2.0 detection and suppresses documentation.
func TestRepoLicenseStartsOnLineOne(t *testing.T) {
	f, err := os.Open("LICENSE")
	if err != nil {
		t.Fatalf("open LICENSE: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("LICENSE is empty")
	}
	firstLine := scanner.Text()
	if strings.TrimSpace(firstLine) == "" {
		t.Error("LICENSE has a leading blank line -- pkg.go.dev will not detect the license;\n" +
			"Fix: remove all blank lines before the license header")
	}
}

// TestRepoLicenseApache20Canonical verifies that the LICENSE file contains
// every structural landmark that github.com/google/licensecheck (used by
// pkg.go.dev) requires to identify Apache-2.0. The scanner is word-based and
// case-insensitive, but all canonical section anchors must be present.
//
// Regression test for: pkg.go.dev reporting "Unknown license" because the
// APPENDIX section was missing from the file after "END OF TERMS AND CONDITIONS".
//
// Required landmarks (in order):
//
//  1. Apache License header with version and URL
//  2. END OF TERMS AND CONDITIONS terminator
//  3. APPENDIX boilerplate instructions (added to canonical text after §9)
//  4. Apache LICENSE-2.0 copy URL
//  5. AS IS disclaimer (part of the boilerplate notice)
func TestRepoLicenseApache20Canonical(t *testing.T) {
	data, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	text := string(data)
	upper := strings.ToUpper(text)

	landmarks := []struct {
		name   string
		needle string // checked case-insensitively
		fix    string
	}{
		{
			name:   "Apache License header",
			needle: "APACHE LICENSE",
			fix:    "Add the canonical Apache License header at the top of the file",
		},
		{
			name:   "version 2.0 declaration",
			needle: "VERSION 2.0",
			fix:    "The header must state 'Version 2.0'",
		},
		{
			name:   "Apache license URL in header",
			needle: "HTTP://WWW.APACHE.ORG/LICENSES/",
			fix:    "Add 'http://www.apache.org/licenses/' to the header",
		},
		{
			name:   "END OF TERMS AND CONDITIONS terminator",
			needle: "END OF TERMS AND CONDITIONS",
			fix:    "The canonical Apache-2.0 text ends with 'END OF TERMS AND CONDITIONS'",
		},
		{
			name:   "APPENDIX section",
			needle: "APPENDIX: HOW TO APPLY THE APACHE LICENSE TO YOUR WORK",
			fix: "Add the APPENDIX boilerplate after 'END OF TERMS AND CONDITIONS'.\n" +
				"Without it, pkg.go.dev's licensecheck cannot identify Apache-2.0 and\n" +
				"will suppress documentation for the module.",
		},
		{
			name:   "Apache LICENSE-2.0 copy URL in boilerplate",
			needle: "HTTP://WWW.APACHE.ORG/LICENSES/LICENSE-2.0",
			fix:    "The APPENDIX boilerplate must include the copy URL 'http://www.apache.org/licenses/LICENSE-2.0'",
		},
		{
			name:   "AS IS disclaimer in boilerplate",
			needle: `"AS IS" BASIS`,
			fix:    `The APPENDIX boilerplate must include the '"AS IS" BASIS' disclaimer`,
		},
	}

	for _, lm := range landmarks {
		if !strings.Contains(upper, lm.needle) {
			t.Errorf("LICENSE missing %q landmark\nFix: %s", lm.name, lm.fix)
		}
	}
}

// TestRepoLicenseApache20SectionOrder verifies that the canonical Apache-2.0
// structural landmarks appear in the correct order. An out-of-order or
// duplicated section would also confuse licensecheck.
func TestRepoLicenseApache20SectionOrder(t *testing.T) {
	data, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	upper := strings.ToUpper(string(data))

	ordered := []string{
		"APACHE LICENSE",
		"VERSION 2.0",
		"TERMS AND CONDITIONS FOR USE",
		"END OF TERMS AND CONDITIONS",
		"APPENDIX: HOW TO APPLY THE APACHE LICENSE TO YOUR WORK",
		"HTTP://WWW.APACHE.ORG/LICENSES/LICENSE-2.0",
	}

	prev := 0
	for _, section := range ordered {
		idx := strings.Index(upper[prev:], section)
		if idx < 0 {
			t.Errorf("LICENSE section %q not found after position %d", section, prev)
			continue
		}
		prev += idx + len(section)
	}
}

// TestRepoLicenseFilename verifies the LICENSE file uses the exact filename
// that pkg.go.dev's crawler accepts. The match is case-insensitive on the
// server, but the canonical casing is "LICENSE" (no extension).
func TestRepoLicenseFilename(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Name() == "LICENSE" {
			found = true
			break
		}
	}
	if !found {
		t.Error(`LICENSE file not found in module root.\n` +
			`pkg.go.dev looks for: LICENSE, LICENSE.md, LICENSE.txt, LICENCE, etc.\n` +
			`The canonical Go module filename is "LICENSE" (no extension).`)
	}
}

// TestRepoLicenseNotEmpty verifies the LICENSE file has meaningful content.
// An empty or near-empty file will score too low in licensecheck's coverage
// threshold and be treated as unknown.
func TestRepoLicenseNotEmpty(t *testing.T) {
	info, err := os.Stat("LICENSE")
	if err != nil {
		t.Fatalf("stat LICENSE: %v", err)
	}
	const minBytes = 10_000 // canonical Apache-2.0 text is ~11 KB
	if info.Size() < minBytes {
		t.Errorf("LICENSE is only %d bytes (want >= %d).\n"+
			"The full Apache-2.0 canonical text is required for licensecheck detection.",
			info.Size(), minBytes)
	}
}

// TestRepoLicenseApache20ExactPhrases verifies that key phrases in the LICENSE
// file match the official Apache-2.0 text word-for-word. licensecheck uses
// word-based LRE matching; even a single extra or substituted word (e.g.,
// "the Licensor" vs "Licensor", "any notices" vs "those notices") can drop
// the coverage score below the detection threshold.
//
// Reference: https://www.apache.org/licenses/LICENSE-2.0.txt
//
// Regression test for: pkg.go.dev reporting "UNKNOWN" license due to three
// word-level deviations from the canonical text.
func TestRepoLicenseApache20ExactPhrases(t *testing.T) {
	data, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	text := string(data)

	// Each phrase is copied verbatim from the official Apache-2.0 text.
	// If the LICENSE file has a word-level deviation, the test will fail
	// and identify the exact phrase that needs correction.
	exactPhrases := []struct {
		canonical string
		badAlt    string // common mistake to watch for
		desc      string
	}{
		{
			canonical: "submitted to Licensor for inclusion in the Work",
			badAlt:    "submitted to the Licensor for inclusion",
			desc:      "extra 'the' before 'Licensor' in Contribution definition",
		},
		{
			canonical: "has been received by Licensor and",
			badAlt:    "has been received by the Licensor and",
			desc:      "extra 'the' before 'Licensor' in Contributor definition",
		},
		{
			canonical: "excluding those notices that do not",
			badAlt:    "excluding any notices that do not",
			desc:      "'any' used instead of canonical 'those' in s4(d)",
		},
		{
			canonical: "submitted to Licensor for inclusion in the Work by the copyright owner",
			badAlt:    "",
			desc:      "full Contribution 'submitted' clause",
		},
		{
			canonical: "on behalf of whom a Contribution has been received by Licensor and",
			badAlt:    "",
			desc:      "full Contributor definition clause",
		},
	}

	for _, p := range exactPhrases {
		if !strings.Contains(text, p.canonical) {
			msg := "LICENSE missing canonical phrase: %q\nDescription: %s\n"
			if p.badAlt != "" && strings.Contains(text, p.badAlt) {
				msg += "Found incorrect variant: %q\n"
				msg += "Fix: replace with the exact text from https://www.apache.org/licenses/LICENSE-2.0.txt"
				t.Errorf(msg, p.canonical, p.desc, p.badAlt)
			} else {
				msg += "Fix: compare against https://www.apache.org/licenses/LICENSE-2.0.txt"
				t.Errorf(msg, p.canonical, p.desc)
			}
		}
	}
}
