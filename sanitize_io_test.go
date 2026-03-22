package demojify_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

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
		{
			// AllowedEmojis preserves one codepoint; only the other is removed.
			// EmojiRemoved must reflect actual removal, not input count.
			name:  "allowed emoji preserved -- only removed ones counted",
			input: "Hello \U0001F600 and \U0001F680",
			opts: demojify.Options{
				RemoveEmojis:        true,
				NormalizeWhitespace: true,
				AllowedEmojis:       []string{"\U0001F680"},
			},
			wantEmoji:   1,
			wantCleaned: "Hello and \U0001F680",
			wantSaved:   len("Hello \U0001F600 and \U0001F680") - len("Hello and \U0001F680"),
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

	// Lines longer than the old 64 KiB default must now succeed (up to 1 MiB).
	t.Run("line longer than 64KiB succeeds", func(t *testing.T) {
		longLine := strings.Repeat("a", 128*1024) // 128 KiB -- exceeds old limit
		var buf bytes.Buffer
		if err := demojify.SanitizeReader(strings.NewReader(longLine), &buf, demojify.Options{}); err != nil {
			t.Fatalf("unexpected error for 128 KiB line: %v", err)
		}
		if buf.String() != longLine {
			t.Error("128 KiB line was not passed through unchanged")
		}
	})

	// Lines exceeding sanitizeReaderMaxTokenSize (1 MiB) must return bufio.ErrTooLong.
	t.Run("line exceeding 1MiB returns ErrTooLong", func(t *testing.T) {
		tooBig := strings.Repeat("b", 1024*1024+1) // 1 MiB + 1 byte
		var buf bytes.Buffer
		err := demojify.SanitizeReader(strings.NewReader(tooBig), &buf, demojify.Options{})
		if !errors.Is(err, bufio.ErrTooLong) {
			t.Fatalf("expected bufio.ErrTooLong, got %v", err)
		}
	})
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
		{
			// Trailing non-whitespace after a valid value must be rejected.
			name:    "trailing data rejected",
			input:   `{"a":1} trailing`,
			opts:    demojify.DefaultOptions(),
			wantErr: true,
		},
		{
			// Two concatenated JSON values must be rejected.
			name:    "multiple values rejected",
			input:   `{"a":1}{"b":2}`,
			opts:    demojify.DefaultOptions(),
			wantErr: true,
		},
		{
			// Trailing whitespace only is accepted (the decoder skips it).
			name:  "trailing whitespace accepted",
			input: `{"a":1}   `,
			opts:  demojify.DefaultOptions(),
			want:  `{"a":1}`,
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
