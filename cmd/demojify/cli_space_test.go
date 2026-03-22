package main_test

import (
	"os"
	"strings"
	"testing"
)

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
