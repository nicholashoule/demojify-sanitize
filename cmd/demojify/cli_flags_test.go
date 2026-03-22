package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
