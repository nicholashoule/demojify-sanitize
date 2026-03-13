package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testBinary is the path to the compiled CLI binary used by all tests.
var testBinary string

func TestMain(m *testing.M) {
	// Build the CLI binary once for all integration tests.
	tmp, err := os.MkdirTemp("", "demojify-cli-test-*")
	if err != nil {
		panic(err)
	}
	bin := filepath.Join(tmp, "demojify")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build CLI binary: " + err.Error())
	}
	testBinary = bin

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// writeTempFile creates a file inside dir with the given name and content.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// runCLI executes the test binary with the given args and returns stdout,
// stderr, and the exit code. A non-zero exit code is not treated as an error.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(testBinary, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("exec error: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestCleanDirectoryExitZero(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "clean.txt", "No emoji here\n")

	stdout, _, code := runCLI(t, "-root", dir)
	if code != 0 {
		t.Errorf("exit code = %d, want 0 for clean directory", code)
	}
	if !strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout = %q, want [PASS] message", stdout)
	}
}

func TestEmojiFoundExitOne(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 build passed\n")

	stdout, _, code := runCLI(t, "-root", dir)
	if code != 1 {
		t.Errorf("exit code = %d, want 1 when emoji found (audit mode)", code)
	}
	if !strings.Contains(stdout, "[WARN]") {
		t.Errorf("stdout = %q, want [WARN] report", stdout)
	}
}

func TestFixStripsEmoji(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "fix.txt", "\U0001F680 Deploy complete\n")

	stdout, _, code := runCLI(t, "-root", dir, "-fix")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -fix", code)
	}
	if !strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout = %q, want [PASS] confirmation", stdout)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if strings.ContainsAny(string(data), "\U0001F680") {
		t.Errorf("file still contains emoji after -fix: %q", data)
	}
}

func TestSubSubstitutesEmoji(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "sub.txt", "\u2705 tests passed\n")

	stdout, _, code := runCLI(t, "-root", dir, "-sub")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -sub", code)
	}
	if !strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout = %q, want [PASS] confirmation", stdout)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read substituted file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "[PASS]") {
		t.Errorf("file = %q, want checkmark replaced with [PASS]", content)
	}
}

func TestSubNormalize(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "norm.txt", "\u2705 tests  passed\n\n\n\nDone.\n")

	_, _, code := runCLI(t, "-root", dir, "-sub", "-normalize")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -sub -normalize", code)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "\n\n\n") {
		t.Errorf("file still has triple newlines after -normalize: %q", content)
	}
	if strings.Contains(content, "  ") {
		t.Errorf("file still has double spaces after -normalize: %q", content)
	}
}

func TestQuietSuppressesOutput(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 done\n")

	stdout, _, code := runCLI(t, "-root", dir, "-quiet")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (audit mode, emoji found)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty with -quiet", stdout)
	}
}

func TestQuietFixSuppressesOutput(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 done\n")

	stdout, _, code := runCLI(t, "-root", dir, "-fix", "-quiet")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -fix -quiet", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty with -quiet", stdout)
	}
}

func TestExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.md", "\u2705 done\n")
	writeTempFile(t, dir, "emoji.txt", "\u274c fail\n")

	// Only scan .md files; .txt should be ignored.
	stdout, _, code := runCLI(t, "-root", dir, "-exts", ".md")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (emoji in .md)", code)
	}
	if strings.Contains(stdout, "emoji.txt") {
		t.Errorf("stdout mentions emoji.txt but -exts .md should skip it: %s", stdout)
	}
	if !strings.Contains(stdout, "emoji.md") {
		t.Errorf("stdout = %q, want emoji.md reported", stdout)
	}
}

func TestExtensionFilterWithoutDot(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.md", "\u2705 done\n")

	// "md" without leading dot should still work.
	_, _, code := runCLI(t, "-root", dir, "-exts", "md")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (emoji in .md via -exts md)", code)
	}
}

func TestSkipFlag(t *testing.T) {
	root := t.TempDir()
	// Create two subdirectories, each with an emoji file.
	for _, sub := range []string{"keep", "skipme"} {
		d := filepath.Join(root, sub)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		writeTempFile(t, d, "emoji.txt", "\u2705 done\n")
	}

	stdout, _, code := runCLI(t, "-root", root, "-skip", "skipme")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (emoji in keep/)", code)
	}
	if strings.Contains(stdout, "skipme") {
		t.Errorf("stdout mentions skipme but it should be skipped: %s", stdout)
	}
	if !strings.Contains(stdout, "keep/emoji.txt") {
		t.Errorf("stdout = %q, want keep/emoji.txt reported", stdout)
	}
}

func TestSkipFlagWithTrailingSlash(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "dist")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeTempFile(t, d, "emoji.txt", "\u2705 done\n")
	writeTempFile(t, root, "clean.txt", "no emoji\n")

	// Trailing slash already present -- should still work.
	stdout, _, code := runCLI(t, "-root", root, "-skip", "dist/")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (dist/ skipped, root is clean)", code)
	}
	if strings.Contains(stdout, "dist") {
		t.Errorf("stdout mentions dist but it should be skipped: %s", stdout)
	}
}

func TestBadRootExitsNonZero(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir", "deep")
	_, stderr, code := runCLI(t, "-root", missing)
	if code == 0 {
		t.Error("exit code = 0, want non-zero for missing root")
	}
	if !strings.Contains(stderr, "root directory") {
		t.Errorf("stderr = %q, want 'root directory' error", stderr)
	}
}

func TestRootNotDirectoryExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "file.txt", "not a dir\n")

	_, stderr, code := runCLI(t, "-root", path)
	if code == 0 {
		t.Error("exit code = 0, want non-zero when root is a file")
	}
	if !strings.Contains(stderr, "not a directory") {
		t.Errorf("stderr = %q, want 'not a directory' error", stderr)
	}
}

func TestFixIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "idem.txt", "\U0001F680 rocket\n")

	// First fix
	_, _, code := runCLI(t, "-root", dir, "-fix")
	if code != 0 {
		t.Fatalf("first -fix exit code = %d, want 0", code)
	}

	data1, _ := os.ReadFile(path)

	// Second run should find nothing
	stdout, _, code := runCLI(t, "-root", dir)
	if code != 0 {
		t.Errorf("second audit exit code = %d, want 0 (already clean)", code)
	}
	if !strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout = %q, want [PASS] after fix", stdout)
	}

	data2, _ := os.ReadFile(path)
	if !bytes.Equal(data1, data2) {
		t.Errorf("file changed between runs: %q -> %q", data1, data2)
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 for -version", code)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty for -version", stderr)
	}
	if !strings.HasPrefix(stdout, "demojify ") {
		t.Fatalf("stdout = %q, want output beginning with \"demojify \"", stdout)
	}
	// Version token must be non-empty and either a semver tag or "(devel)".
	token := strings.TrimSpace(strings.TrimPrefix(stdout, "demojify "))
	if token == "" {
		t.Error("version token is empty")
	}
	isSemver := strings.HasPrefix(token, "v") && strings.Contains(token, ".")
	isDevel := strings.HasPrefix(token, "(devel)")
	if !isSemver && !isDevel {
		t.Errorf("version token %q is neither a semver tag nor \"(devel)\"", token)
	}
	// Output must end with a newline (fmt.Println guarantee).
	if !strings.HasSuffix(stdout, "\n") {
		t.Errorf("stdout = %q, want trailing newline", stdout)
	}
}

// TestVersionFlagNoScan verifies that -version exits immediately without
// performing a directory scan, even when -root points to a directory that
// contains emoji.
func TestVersionFlagNoScan(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 check\n")

	stdout, _, code := runCLI(t, "-version", "-root", dir)
	if code != 0 {
		t.Errorf("exit code = %d, want 0 for -version", code)
	}
	if strings.Contains(stdout, "[WARN]") || strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout = %q, want no scan output when -version is set", stdout)
	}
}

// TestNestedFilePathForwardSlash verifies that [WARN] output always contains
// forward-slash separators regardless of the host OS. ScanDir normalises paths
// with filepath.ToSlash so Windows callers get "sub/file.txt" rather than
// "sub\file.txt".
func TestNestedFilePathForwardSlash(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeTempFile(t, sub, "nested.txt", "\u2705 done\n")

	stdout, _, code := runCLI(t, "-root", root)
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (emoji found)", code)
	}
	// Path must use forward slashes on every platform.
	if !strings.Contains(stdout, "sub/nested.txt") {
		t.Errorf("stdout = %q, want forward-slash path \"sub/nested.txt\"", stdout)
	}
	if strings.Contains(stdout, `sub\nested.txt`) {
		t.Errorf("stdout = %q, contains backslash path; want forward slash", stdout)
	}
}

func TestJSONCleanDirectory(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "clean.txt", "No emoji here\n")

	stdout, _, code := runCLI(t, "-root", dir, "-json")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	var result struct {
		Findings []json.RawMessage `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout: %s", err, stdout)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}

func TestJSONEmojiFound(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 done\n")

	stdout, _, code := runCLI(t, "-root", dir, "-json")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	var result struct {
		Findings []struct {
			Path     string `json:"path"`
			HasEmoji bool   `json:"hasEmoji"`
			Matches  []struct {
				Sequence string `json:"sequence"`
				Line     int    `json:"line"`
				Column   int    `json:"column"`
				Context  string `json:"context"`
			} `json:"matches"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout: %s", err, stdout)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	f := result.Findings[0]
	if f.Path != "emoji.txt" {
		t.Errorf("path = %q, want %q", f.Path, "emoji.txt")
	}
	if !f.HasEmoji {
		t.Error("hasEmoji = false, want true")
	}
	if len(f.Matches) == 0 {
		t.Fatal("expected at least one match")
	}
	m := f.Matches[0]
	if m.Line != 1 {
		t.Errorf("match line = %d, want 1", m.Line)
	}
	if m.Column != 0 {
		t.Errorf("match column = %d, want 0", m.Column)
	}
	if m.Context == "" {
		t.Error("match context is empty, want full line text")
	}
}

func TestJSONFix(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "fix.txt", "\U0001F680 Deploy complete\n")

	stdout, _, code := runCLI(t, "-root", dir, "-fix", "-json")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -fix -json", code)
	}

	var result struct {
		Findings []struct {
			Fixed *struct {
				Success bool `json:"success"`
				Count   int  `json:"count"`
			} `json:"fixed"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout: %s", err, stdout)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected at least 1 finding")
	}
	if result.Findings[0].Fixed == nil {
		t.Fatal("expected non-nil fixed field")
	}
	if !result.Findings[0].Fixed.Success {
		t.Error("fixed.success = false, want true")
	}

	// Verify the file was actually cleaned.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if strings.ContainsAny(string(data), "\U0001F680") {
		t.Errorf("file still contains emoji after -fix -json: %q", data)
	}
}

func TestJSONSub(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "sub.txt", "\u2705 tests passed\n")

	stdout, _, code := runCLI(t, "-root", dir, "-sub", "-json")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 after -sub -json", code)
	}

	var result struct {
		Findings []struct {
			Matches []struct {
				Replacement string `json:"replacement"`
			} `json:"matches"`
			Fixed *struct {
				Success bool `json:"success"`
			} `json:"fixed"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout: %s", err, stdout)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected at least 1 finding")
	}
	f := result.Findings[0]
	if len(f.Matches) == 0 {
		t.Fatal("expected at least 1 match")
	}
	if f.Matches[0].Replacement == "" {
		t.Error("expected non-empty replacement for checkmark with -sub")
	}
	if f.Fixed == nil || !f.Fixed.Success {
		t.Error("expected successful fix with -sub")
	}
}

func TestJSONOverridesQuiet(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 done\n")

	stdout, _, code := runCLI(t, "-root", dir, "-json", "-quiet")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	// -json overrides -quiet: JSON output should still appear.
	var result struct {
		Findings []json.RawMessage `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("-json -quiet produced no JSON: %v\nstdout: %s", err, stdout)
	}
	if len(result.Findings) == 0 {
		t.Error("expected findings in JSON output")
	}
}

func TestJSONNoTextLeakage(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "emoji.txt", "\u2705 done\n")

	stdout, _, _ := runCLI(t, "-root", dir, "-json")
	// JSON mode must not emit human-readable markers.
	if strings.Contains(stdout, "[WARN]") {
		t.Errorf("stdout contains [WARN] in -json mode: %s", stdout)
	}
	if strings.Contains(stdout, "[PASS]") {
		t.Errorf("stdout contains [PASS] in -json mode: %s", stdout)
	}
}

// TestFixCollapseSpacesAfterEmojiRemoval verifies that -fix not only strips
// emoji but also collapses any inline double-spaces and trailing spaces that
// are left behind as artifacts of the removal. This covers the downstream-
// consumer issue where lines like
//
//	"# ⚠️⚠️  The text here. "
//
// were being written back as "#   The text here. " rather than
// "# The text here.".
func TestFixCollapseSpacesAfterEmojiRemoval(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "markdown heading double warning emoji inline spaces trailing space",
			// "# ⚠️⚠️  The text here. " with trailing space
			input: "# \u26A0\uFE0F\u26A0\uFE0F  The text here. \n",
			want:  "# The text here.\n",
		},
		{
			name:  "single warning emoji with extra spaces",
			input: "# \u26A0\uFE0F  A section title. \n",
			want:  "# A section title.\n",
		},
		{
			name:  "bullet point double checkmark with trailing space",
			input: "- \u2705\u2705  Both checks passed. \n",
			want:  "- Both checks passed.\n",
		},
		{
			name:  "multiple emoji separated by text inline spaces",
			input: "NOTE: \u26A0\uFE0F  Check  \u26A0\uFE0F  logs. \n",
			want:  "NOTE: Check logs.\n",
		},
		{
			name:  "multi-line file with emoji in heading and clean body",
			input: "# \u26A0\uFE0F  Warning title. \n\nBody text here.\n",
			want:  "# Warning title.\n\nBody text here.\n",
		},
		{
			name:  "three consecutive spaces left after emoji removal",
			input: "Item: \u2705\u2705\u2705   All done. \n",
			want:  "Item: All done.\n",
		},
		{
			name:  "emoji adjacent to punctuation with extra space",
			input: "Status: \u274C  Failed. \n",
			want:  "Status: Failed.\n",
		},
		{
			name:  "multiple lines each with inline emoji and spaces",
			input: "# \u26A0\uFE0F  Line one. \n# \u2705  Line two. \n",
			want:  "# Line one.\n# Line two.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeTempFile(t, dir, "fix.txt", tt.input)

			_, _, code := runCLI(t, "-root", dir, "-fix")
			if code != 0 {
				t.Fatalf("exit code = %d, want 0 after -fix", code)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixed file: %v", err)
			}
			got := string(data)
			if got != tt.want {
				t.Errorf("-fix result:\n  got  %q\n  want %q", got, tt.want)
			}
		})
	}
}

// TestSubCollapseSpacesAfterSubstitution verifies that -sub collapses inline
// double-spaces and trailing spaces that arise after emoji are replaced with
// text tokens. For example
//
//	"# ⚠️  The text here. "
//
// must become "# WARNING The text here." rather than leaving extra spaces.
//
// Consecutive identical tokens produced by adjacent repeated emoji are also
// collapsed to a single token: "⚠️⚠️" -> "WARNING" (not "WARNINGWARNING").
// Tokens that differ (e.g. [PASS] followed by [FAIL]) are preserved as-is.
func TestSubCollapseSpacesAfterSubstitution(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			// Two adjacent emoji (no space between) -> single collapsed token.
			name:  "heading two adjacent warning emoji collapsed to one token",
			input: "# \u26A0\uFE0F\u26A0\uFE0F  The text here. \n",
			want:  "# WARNING The text here.\n",
		},
		{
			// Two emoji separated by a space -> single collapsed token.
			name:  "heading two spaced warning emoji collapsed to one token",
			input: "# \u26A0\uFE0F \u26A0\uFE0F  The text here. \n",
			want:  "# WARNING The text here.\n",
		},
		{
			name:  "single warning emoji token with extra spaces",
			input: "# \u26A0\uFE0F  A section title. \n",
			want:  "# WARNING A section title.\n",
		},
		{
			name:  "checkmark token with trailing space",
			input: "Build: \u2705  Passed. \n",
			want:  "Build: [PASS] Passed.\n",
		},
		{
			name:  "fail token with inline spaces",
			input: "Result: \u274C  Error  occurred. \n",
			want:  "Result: [FAIL] Error occurred.\n",
		},
		{
			// Emoji with space between them -> tokens with space between.
			name:  "mixed spaced tokens in one line",
			input: "# \u2705 \u274C  Mixed outcome. \n",
			want:  "# [PASS] [FAIL] Mixed outcome.\n",
		},
		{
			// Adjacent emoji (no space) -> concatenated tokens.
			name:  "adjacent checkmark and x-mark tokens concatenated",
			input: "# \u2705\u274C  Mixed outcome. \n",
			want:  "# [PASS][FAIL] Mixed outcome.\n",
		},
		{
			name:  "multi-line with tokens",
			input: "# \u26A0\uFE0F  Warning title. \n\nBody text.\n",
			want:  "# WARNING Warning title.\n\nBody text.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeTempFile(t, dir, "sub.txt", tt.input)

			_, _, code := runCLI(t, "-root", dir, "-sub")
			if code != 0 {
				t.Fatalf("exit code = %d, want 0 after -sub", code)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read substituted file: %v", err)
			}
			got := string(data)
			if got != tt.want {
				t.Errorf("-sub result:\n  got  %q\n  want %q", got, tt.want)
			}
		})
	}
}

// TestFixSubIdempotentAfterSpaceCleanup verifies that running -fix or -sub a
// second time after the first pass reports no further changes, because the
// inline-space cleanup from the first pass already produced a clean file.
func TestFixSubIdempotentAfterSpaceCleanup(t *testing.T) {
	// Use a single warning emoji so that -sub produces "# WARNING The text"
	// rather than a concatenated token, keeping the idempotency assertion clean.
	input := "# \u26A0\uFE0F  The text here. \n"

	for _, flag := range []string{"-fix", "-sub"} {
		t.Run(flag, func(t *testing.T) {
			dir := t.TempDir()
			writeTempFile(t, dir, "idem.txt", input)

			// First pass: should find and fix emoji.
			_, _, code := runCLI(t, "-root", dir, flag)
			if code != 0 {
				t.Fatalf("first %s exit code = %d, want 0", flag, code)
			}

			// Second pass (audit): should report clean.
			stdout, _, code := runCLI(t, "-root", dir)
			if code != 0 {
				t.Errorf("second audit after %s: exit code = %d, want 0 (file should be clean)", flag, code)
			}
			if !strings.Contains(stdout, "[PASS]") {
				t.Errorf("second audit after %s: stdout = %q, want [PASS]", flag, stdout)
			}
		})
	}
}
