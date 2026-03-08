package demojify_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	t.Run("CRLF file is sanitized and line endings are normalized to LF", func(t *testing.T) {
		// SanitizeFile passes content through Normalize which converts CRLF to LF.
		// Verify the written file contains LF-only endings on every platform.
		dir := t.TempDir()
		path := writeTempFile(t, dir, "crlf.txt", "\u2705 passed\r\nDone.\r\n")
		changed, err := demojify.SanitizeFile(path, demojify.DefaultOptions())
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
		content := string(data)
		if strings.Contains(content, "\r") {
			t.Errorf("file still contains CR after SanitizeFile: %q", content)
		}
		if strings.Contains(content, "\u2705") {
			t.Errorf("file still contains emoji after SanitizeFile: %q", content)
		}
	})

	t.Run("binary file is skipped", func(t *testing.T) {
		dir := t.TempDir()
		// Write a file containing a NUL byte -- detected as binary.
		binPath := filepath.Join(dir, "binary.bin")
		if err := os.WriteFile(binPath, []byte("prefix\x00suffix \u2705"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		origStat, _ := os.Stat(binPath)
		changed, err := demojify.SanitizeFile(binPath, demojify.DefaultOptions())
		if err != nil {
			t.Fatalf("SanitizeFile on binary: %v", err)
		}
		if changed {
			t.Error("changed = true for binary file, want false")
		}
		// File must be untouched.
		newStat, _ := os.Stat(binPath)
		if !newStat.ModTime().Equal(origStat.ModTime()) {
			t.Error("binary file was modified by SanitizeFile")
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

func TestSanitizeReport(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		opts        demojify.Options
		wantEmoji   int
		wantSaved   int
		wantCleaned string
	}{
		{
			name:        "emoji removed and counted",
			input:       "Hello \U0001F600 World \U0001F680",
			opts:        demojify.DefaultOptions(),
			wantEmoji:   2,
			wantSaved:   10,
			wantCleaned: "Hello World",
		},
		{
			name:        "no emoji",
			input:       "Hello World",
			opts:        demojify.DefaultOptions(),
			wantEmoji:   0,
			wantSaved:   0,
			wantCleaned: "Hello World",
		},
		{
			name:        "emoji removal disabled",
			input:       "Hello \U0001F600",
			opts:        demojify.Options{NormalizeWhitespace: true},
			wantEmoji:   0,
			wantSaved:   0,
			wantCleaned: "Hello \U0001F600",
		},
		{
			name:        "whitespace normalization adds savings",
			input:       "\U0001F680  deploy\n\n\nstatus",
			opts:        demojify.DefaultOptions(),
			wantEmoji:   1,
			wantSaved:   len("\U0001F680  deploy\n\n\nstatus") - len("deploy\n\nstatus"),
			wantCleaned: "deploy\n\nstatus",
		},
		{
			name:        "empty input",
			input:       "",
			opts:        demojify.DefaultOptions(),
			wantEmoji:   0,
			wantSaved:   0,
			wantCleaned: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := demojify.SanitizeReport(tt.input, tt.opts)
			if result.Cleaned != tt.wantCleaned {
				t.Errorf("Cleaned = %q, want %q", result.Cleaned, tt.wantCleaned)
			}
			if result.EmojiRemoved != tt.wantEmoji {
				t.Errorf("EmojiRemoved = %d, want %d", result.EmojiRemoved, tt.wantEmoji)
			}
			if result.BytesSaved != tt.wantSaved {
				t.Errorf("BytesSaved = %d, want %d", result.BytesSaved, tt.wantSaved)
			}
		})
	}
}

func TestSanitizeReader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		opts  demojify.Options
		want  string
	}{
		{
			name:  "emoji removal only",
			input: "Hello \U0001F600 World\nanother line",
			opts:  demojify.Options{RemoveEmojis: true},
			want:  "Hello  World\nanother line",
		},
		{
			name:  "emoji removal with normalization",
			input: "Hello \U0001F600 World\nanother line",
			opts:  demojify.DefaultOptions(),
			want:  "Hello World\nanother line",
		},
		{
			name:  "collapse blank lines",
			input: "hello\n\n\n\nworld",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "hello\n\nworld",
		},
		{
			name:  "skip leading blank lines",
			input: "\n\nhello\nworld",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "hello\nworld",
		},
		{
			name:  "skip trailing blank lines",
			input: "hello\nworld\n\n\n",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "hello\nworld",
		},
		{
			name:  "no options passthrough",
			input: "Hello \U0001F600  World\n\n\n",
			opts:  demojify.Options{},
			want:  "Hello \U0001F600  World\n\n",
		},
		{
			name:  "empty input",
			input: "",
			opts:  demojify.DefaultOptions(),
			want:  "",
		},
		{
			name:  "single line no newline",
			input: "Hello \U0001F680",
			opts:  demojify.Options{RemoveEmojis: true},
			want:  "Hello ",
		},
		{
			name:  "preserve leading indentation",
			input: "  indented \U0001F600 text",
			opts:  demojify.DefaultOptions(),
			want:  "  indented text",
		},
		{
			name:  "CRLF handled by scanner",
			input: "line1\r\nline2\r\n",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := demojify.SanitizeReader(strings.NewReader(tt.input), &buf, tt.opts)
			if err != nil {
				t.Fatalf("SanitizeReader error: %v", err)
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("SanitizeReader:\ngot  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    demojify.Options
		want    string
		wantErr bool
	}{
		{
			name:  "string values sanitized",
			input: `{"message":"Hello \u2705 World","count":42}`,
			opts:  demojify.Options{RemoveEmojis: true},
			want:  `{"count":42,"message":"Hello  World"}`,
		},
		{
			name:  "nested objects",
			input: `{"outer":{"inner":"test \u2705"}}`,
			opts:  demojify.Options{RemoveEmojis: true},
			want:  `{"outer":{"inner":"test "}}`,
		},
		{
			name:  "arrays",
			input: `["Hello \u2705","World \u274c"]`,
			opts:  demojify.Options{RemoveEmojis: true},
			want:  `["Hello ","World "]`,
		},
		{
			name:  "keys preserved",
			input: `{"status":"done \u2705"}`,
			opts:  demojify.Options{RemoveEmojis: true},
			want:  `{"status":"done "}`,
		},
		{
			name:  "numbers preserved",
			input: `{"val":123456789012345678}`,
			opts:  demojify.DefaultOptions(),
			want:  `{"val":123456789012345678}`,
		},
		{
			name:  "booleans and null preserved",
			input: `{"flag":true,"empty":null}`,
			opts:  demojify.DefaultOptions(),
			want:  `{"empty":null,"flag":true}`,
		},
		{
			name:    "invalid JSON",
			input:   `{not json}`,
			opts:    demojify.DefaultOptions(),
			wantErr: true,
		},
		{
			name:  "empty object",
			input: `{}`,
			opts:  demojify.DefaultOptions(),
			want:  `{}`,
		},
		{
			name:  "whitespace normalization in values",
			input: `{"msg":"hello   world\n\n\ntoo many lines"}`,
			opts:  demojify.DefaultOptions(),
			want:  `{"msg":"hello world\n\ntoo many lines"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := demojify.SanitizeJSON([]byte(tt.input), tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("SanitizeJSON error: %v", err)
			}
			// Normalize by re-marshaling the expected JSON for comparison
			// since map key order is non-deterministic.
			var wantVal, gotVal interface{}
			if err := json.Unmarshal([]byte(tt.want), &wantVal); err != nil {
				t.Fatalf("unmarshal want: %v", err)
			}
			if err := json.Unmarshal(got, &gotVal); err != nil {
				t.Fatalf("unmarshal got: %v", err)
			}
			wantBytes, _ := json.Marshal(wantVal)
			gotBytes, _ := json.Marshal(gotVal)
			if !bytes.Equal(gotBytes, wantBytes) {
				t.Errorf("SanitizeJSON:\ngot  %s\nwant %s", got, tt.want)
			}
		})
	}
}

// TestSanitizeFileWhitespaceOnlyChanges verifies that SanitizeFile writes
// back the file when only whitespace normalization changes the content
// (no emoji present). This is the edge case where RemoveEmojis produces
// no change but NormalizeWhitespace does.
func TestSanitizeFileWhitespaceOnlyChanges(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "spaces.txt", "hello   world\n\n\n\nend\n")

	opts := demojify.DefaultOptions() // both RemoveEmojis and NormalizeWhitespace
	changed, err := demojify.SanitizeFile(path, opts)
	if err != nil {
		t.Fatalf("SanitizeFile: %v", err)
	}
	if !changed {
		t.Error("changed = false, want true for whitespace-only normalization")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "   ") {
		t.Errorf("file still contains triple spaces: %q", content)
	}
	if strings.Contains(content, "\n\n\n") {
		t.Errorf("file still contains triple newlines: %q", content)
	}
}

// TestSanitizeConcurrent verifies that Sanitize is safe for concurrent use
// from multiple goroutines. The race detector (go test -race) will catch
// any data races on the compiled package-level regexes.
func TestSanitizeConcurrent(t *testing.T) {
	const goroutines = 50
	inputs := []struct {
		text string
		opts demojify.Options
	}{
		{"\U0001F680 Deploy complete!\n\n\nCheck the dashboard \U0001F4CA", demojify.DefaultOptions()},
		{"Hello   World\n\n\nMore text", demojify.Options{NormalizeWhitespace: true}},
		{"No changes needed", demojify.Options{}},
		{"\u2705 pass \u274C fail", demojify.Options{RemoveEmojis: true}},
		{"", demojify.DefaultOptions()},
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			in := inputs[idx%len(inputs)]
			result := demojify.Sanitize(in.text, in.opts)
			// Verify the result is at least deterministic.
			if result != demojify.Sanitize(in.text, in.opts) {
				t.Errorf("Sanitize produced non-deterministic result for input %d", idx)
			}
		}(i)
	}
	wg.Wait()
}
