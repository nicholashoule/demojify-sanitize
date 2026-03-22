package main_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

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
