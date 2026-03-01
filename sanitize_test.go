package demojify_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDefaultOptions(t *testing.T) {
	opts := demojify.DefaultOptions()
	if !opts.RemoveEmojis {
		t.Error("DefaultOptions().RemoveEmojis should be true")
	}
	if !opts.NormalizeWhitespace {
		t.Error("DefaultOptions().NormalizeWhitespace should be true")
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		opts  demojify.Options
		want  string
	}{
		{
			name:  "zero options – nothing changed",
			input: "Certainly! 😀 Hello  World",
			opts:  demojify.Options{},
			want:  "Certainly! 😀 Hello  World",
		},
		{
			name:  "remove emojis only",
			input: "Certainly! 😀 Hello  World",
			opts:  demojify.Options{RemoveEmojis: true},
			want:  "Certainly!  Hello  World",
		},
		{
			name:  "normalize whitespace only",
			input: "Hello  World\n\n\nMore text",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "Hello World\n\nMore text",
		},
		{
			name:  "all options – emoji removal and normalization",
			input: "\U0001F680 Deploy complete!\n\n\nCheck the dashboard \U0001F4CA",
			opts:  demojify.DefaultOptions(),
			want:  "Deploy complete!\n\nCheck the dashboard",
		},
		{
			name:  "all options – multi-space collapsed after emoji removal",
			input: "\U0001F680  double space",
			opts:  demojify.DefaultOptions(),
			want:  "double space",
		},
		{
			name:  "AllowedRanges – preserve rocket, remove bar chart",
			input: "Deploy \U0001F680 done. Check \U0001F4CA.",
			opts: demojify.Options{
				RemoveEmojis: true,
				AllowedRanges: []*unicode.RangeTable{
					{R32: []unicode.Range32{{Lo: 0x1F680, Hi: 0x1F680, Stride: 1}}},
				},
			},
			want: "Deploy \U0001F680 done. Check .",
		},
		{
			name:  "AllowedRanges nil – behaves identically to Demojify",
			input: "Hello \U0001F600 World",
			opts:  demojify.Options{RemoveEmojis: true, AllowedRanges: nil},
			want:  "Hello  World",
		},
		{
			name:  "AllowedEmojis – preserve rocket, remove grinning face",
			input: "Deploy \U0001F680 done \U0001F600!",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{"\U0001F680"},
			},
			want: "Deploy \U0001F680 done !",
		},
		{
			name:  "AllowedEmojis – preserve multiple emoji",
			input: "\U0001F680 Deploy \U0001F4CA chart \U0001F600 smile",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{"\U0001F680", "\U0001F4CA"},
			},
			want: "\U0001F680 Deploy \U0001F4CA chart  smile",
		},
		{
			name:  "AllowedEmojis with AllowedRanges combined",
			input: "Deploy \U0001F680 check \u2705 stars \u2B50",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{"\U0001F680"},
				AllowedRanges: []*unicode.RangeTable{
					{R16: []unicode.Range16{{Lo: 0x2705, Hi: 0x2705, Stride: 1}}},
				},
			},
			want: "Deploy \U0001F680 check \u2705 stars ",
		},
		{
			name:  "AllowedEmojis nil – behaves identically to Demojify",
			input: "Hello \U0001F600 World",
			opts:  demojify.Options{RemoveEmojis: true, AllowedEmojis: nil},
			want:  "Hello  World",
		},
		{
			name:  "AllowedEmojis empty slice – behaves identically to Demojify",
			input: "Hello \U0001F600 World",
			opts:  demojify.Options{RemoveEmojis: true, AllowedEmojis: []string{}},
			want:  "Hello  World",
		},
		{
			name:  "AllowedEmojis – placeholder collision in input text",
			input: "\uFDD00\uFDD0 keep \U0001F680 remove \U0001F600",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{"\U0001F680"},
			},
			want: "\uFDD00\uFDD0 keep \U0001F680 remove ",
		},
		{
			name:  "AllowedEmojis – empty string entry is silently ignored",
			input: "Hello \U0001F600 World",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{""},
			},
			want: "Hello  World",
		},
		{
			name:  "AllowedEmojis – empty string mixed with valid entry",
			input: "Deploy \U0001F680 done \U0001F600!",
			opts: demojify.Options{
				RemoveEmojis:  true,
				AllowedEmojis: []string{"", "\U0001F680"},
			},
			want: "Deploy \U0001F680 done !",
		},
		{
			name:  "empty string",
			input: "",
			opts:  demojify.DefaultOptions(),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Sanitize(tt.input, tt.opts)
			if got != tt.want {
				t.Errorf("Sanitize(%q, %+v)\n  got  %q\n  want %q", tt.input, tt.opts, got, tt.want)
			}
		})
	}
}

func TestSanitizeFile(t *testing.T) {
	t.Run("file with emoji is sanitized and written back", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "dirty.txt", "deploy \U0001F680 done\n")
		changed, err := demojify.SanitizeFile(path, demojify.Options{RemoveEmojis: true})
		if err != nil {
			t.Fatalf("SanitizeFile: %v", err)
		}
		if !changed {
			t.Error("changed = false, want true")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		want := "deploy  done\n"
		if string(data) != want {
			t.Errorf("file content = %q, want %q", string(data), want)
		}
	})

	t.Run("clean file returns false and is not written", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "clean.txt", "no emoji here")
		info1, _ := os.Stat(path)
		changed, err := demojify.SanitizeFile(path, demojify.DefaultOptions())
		if err != nil {
			t.Fatalf("SanitizeFile: %v", err)
		}
		if changed {
			t.Error("changed = true, want false for clean file")
		}
		info2, _ := os.Stat(path)
		if !info2.ModTime().Equal(info1.ModTime()) {
			t.Error("clean file was modified unexpectedly")
		}
	})

	t.Run("file permissions are preserved", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "perms.txt")
		if err := os.WriteFile(path, []byte("check \u2705\n"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		infoBefore, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat before: %v", err)
		}
		wantPerm := infoBefore.Mode().Perm()
		if _, err := demojify.SanitizeFile(path, demojify.Options{RemoveEmojis: true}); err != nil {
			t.Fatalf("SanitizeFile: %v", err)
		}
		infoAfter, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat after: %v", err)
		}
		if infoAfter.Mode().Perm() != wantPerm {
			t.Errorf("permissions changed: got %o, want %o", infoAfter.Mode().Perm(), wantPerm)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir", "no-file.txt")
		_, err := demojify.SanitizeFile(missing, demojify.DefaultOptions())
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}

// TestSanitizeAgentOutputRemediation proves that the module detects and fully
// remediates emoji in AI-generated content. ContainsEmoji catches the
// violation, Sanitize removes emoji in one call, and the result is idempotent.
func TestSanitizeAgentOutputRemediation(t *testing.T) {
	// Simulate AI-generated content with decorative emoji mixed into real text.
	rogueOutput := "\U0001F680 Deployment\n" +
		"\n" +
		"Run the following command to deploy:\n" +
		"\n" +
		"    go build ./...\n" +
		"\n" +
		"Check the docs \U0001F4CA for details."

	// ContainsEmoji detects the violation before the file is written.
	if !demojify.ContainsEmoji(rogueOutput) {
		t.Fatal("ContainsEmoji: expected true for rogue agent output, got false")
	}

	// Sanitize remediates emoji in one call.
	clean := demojify.Sanitize(rogueOutput, demojify.DefaultOptions())

	// Output is now emoji-free.
	if demojify.ContainsEmoji(clean) {
		t.Errorf("after Sanitize, output still contains emoji:\n%s", clean)
	}

	// Substantive content is preserved.
	for _, required := range []string{"Deployment", "go build ./..."} {
		if !strings.Contains(clean, required) {
			t.Errorf("after Sanitize, output is missing expected content %q:\n%s", required, clean)
		}
	}

	// Running Sanitize again produces identical output -- idempotent.
	if twice := demojify.Sanitize(clean, demojify.DefaultOptions()); twice != clean {
		t.Errorf("Sanitize is not idempotent:\nfirst:  %q\nsecond: %q", clean, twice)
	}
}
