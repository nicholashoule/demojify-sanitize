package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
