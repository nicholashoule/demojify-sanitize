package demojify_test

import (
	"os"
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestFindAll(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no emoji returns nil",
			input: "Hello, World!",
			want:  nil,
		},
		{
			name:  "empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:  "single emoji once",
			input: "Deploy \U0001F680 done",
			want:  []string{"\U0001F680"},
		},
		{
			name:  "duplicate emoji appears once",
			input: "\U0001F680 first \U0001F680 second \U0001F680 third",
			want:  []string{"\U0001F680"},
		},
		{
			name:  "multiple distinct emoji preserved in first-occurrence order",
			input: "done \u2705 error \u274c warn \u26a0",
			want:  []string{"\u2705", "\u274c", "\u26a0"},
		},
		{
			name:  "variation selector treated as its own codepoint",
			input: "star \u2b50\ufe0f end",
			want:  []string{"\u2b50", "\ufe0f"},
		},
		{
			name:  "non-emoji unicode is ignored",
			input: "Chinese \u4e2d\u6587 Arabic \u0639\u0631\u0628\u064a",
			want:  nil,
		},
		{
			name:  "supplementary emoji block",
			input: "tip \U0001f4a1 and rocket \U0001F680 and tip again \U0001f4a1",
			want:  []string{"\U0001f4a1", "\U0001F680"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.FindAll(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("FindAll(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i, seq := range got {
				if seq != tt.want[i] {
					t.Errorf("FindAll(%q)[%d] = %q, want %q", tt.input, i, seq, tt.want[i])
				}
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		replacements map[string]string
		want         string
	}{
		{
			name:         "nil map behaves like Demojify",
			input:        "Deploy \U0001F680 done",
			replacements: nil,
			want:         "Deploy  done",
		},
		{
			name:         "empty map behaves like Demojify",
			input:        "Deploy \U0001F680 done",
			replacements: map[string]string{},
			want:         "Deploy  done",
		},
		{
			name:         "mapped codepoint is substituted",
			input:        "status \u2705 ok",
			replacements: map[string]string{"\u2705": "[PASS]"},
			want:         "status [PASS] ok",
		},
		{
			name:         "unmapped emoji is stripped",
			input:        "rocket \U0001F680 here",
			replacements: map[string]string{"\u2705": "[PASS]"},
			want:         "rocket  here",
		},
		{
			name:         "multiple mappings applied",
			input:        "\u2705 pass \u274c fail \u26a0 warn",
			replacements: map[string]string{"\u2705": "[PASS]", "\u274c": "[FAIL]", "\u26a0": "WARNING"},
			want:         "[PASS] pass [FAIL] fail WARNING warn",
		},
		{
			name:         "longer key matched before shorter sub-sequence",
			input:        "\u26a0\ufe0f fire",
			replacements: map[string]string{"\u26a0\ufe0f": "WARNING", "\u26a0": "WARN"},
			want:         "WARNING fire",
		},
		{
			name:         "no emoji in input returns input unchanged",
			input:        "plain text",
			replacements: map[string]string{"\u2705": "[PASS]"},
			want:         "plain text",
		},
		{
			name:         "arrow substitution (non-emoji codepoint in map)",
			input:        "step \u2192 next",
			replacements: map[string]string{"\u2192": "->"},
			want:         "step -> next",
		},
		{
			name:         "map replacement not introducing double spaces",
			input:        "a\u2705b",
			replacements: map[string]string{"\u2705": "[PASS]"},
			want:         "a[PASS]b",
		},
		{
			name:         "empty string",
			input:        "",
			replacements: map[string]string{"\u2705": "[PASS]"},
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Replace(tt.input, tt.replacements)
			if got != tt.want {
				t.Errorf("Replace(%q, ...) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReplaceFile(t *testing.T) {
	t.Run("file with mapped emoji is substituted and written back", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		original := "build \u2705 passed\n"
		if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		repl := map[string]string{"\u2705": "[PASS]"}
		count, err := demojify.ReplaceFile(path, repl)
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count == 0 {
			t.Error("count = 0, want > 0")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile after ReplaceFile: %v", err)
		}
		want := "build [PASS] passed\n"
		if string(data) != want {
			t.Errorf("file content = %q, want %q", string(data), want)
		}
	})

	t.Run("clean file returns zero count and is not written", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "clean.txt")
		content := "no emoji here\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		// Record mtime before the call.
		info1, _ := os.Stat(path)
		count, err := demojify.ReplaceFile(path, map[string]string{"\u2705": "[PASS]"})
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
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
		// Record permissions as set by the OS (Windows may normalise 0600 -> 0666).
		infoBeore, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat before: %v", err)
		}
		wantPerm := infoBeore.Mode().Perm()
		if _, err := demojify.ReplaceFile(path, map[string]string{"\u2705": "[PASS]"}); err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		infoAfter, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat after: %v", err)
		}
		if infoAfter.Mode().Perm() != wantPerm {
			t.Errorf("permissions changed: got %o, want %o", infoAfter.Mode().Perm(), wantPerm)
		}
	})

	t.Run("nil map behaves like Demojify on file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "demojify.txt")
		if err := os.WriteFile(path, []byte("rocket \U0001F680\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		count, err := demojify.ReplaceFile(path, nil)
		if err != nil {
			t.Fatalf("ReplaceFile: %v", err)
		}
		if count == 0 {
			t.Error("count = 0, want > 0")
		}
		data, _ := os.ReadFile(path)
		if string(data) != "rocket \n" {
			t.Errorf("file content = %q, want %q", string(data), "rocket \n")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := demojify.ReplaceFile("/nonexistent/path/no-file.txt", nil)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}

func TestFindAllMapped(t *testing.T) {
	repl := map[string]string{
		"\u2705":       "[PASS]",
		"\u274c":       "[FAIL]",
		"\u26a0":       "WARNING",
		"\u26a0\ufe0f": "WARNING",
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no mapped emoji returns nil",
			input: "plain text",
			want:  nil,
		},
		{
			name:  "empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:  "single mapped key found",
			input: "\u2705 build passed",
			want:  []string{"\u2705"},
		},
		{
			name:  "multiple mapped keys in first-occurrence order",
			input: "\u2705 pass then \u274c fail then \u26a0 warn",
			want:  []string{"\u2705", "\u274c", "\u26a0"},
		},
		{
			name:  "duplicate key appears once",
			input: "\u2705 first \u2705 second \u2705 third",
			want:  []string{"\u2705"},
		},
		{
			name:  "unmapped emoji not in result",
			input: "\U0001F680 rocket \u2705 check",
			want:  []string{"\u2705"},
		},
		{
			name:  "variation selector key matched when present",
			input: "\u26a0\ufe0f critical",
			want:  []string{"\u26a0\ufe0f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.FindAllMapped(tt.input, repl)
			if len(got) != len(tt.want) {
				t.Fatalf("FindAllMapped(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i, seq := range got {
				if seq != tt.want[i] {
					t.Errorf("FindAllMapped(%q)[%d] = %q, want %q", tt.input, i, seq, tt.want[i])
				}
			}
		})
	}
}

func TestFindAllMappedNilMap(t *testing.T) {
	got := demojify.FindAllMapped("hello \u2705", nil)
	if len(got) != 0 {
		t.Errorf("FindAllMapped with nil map = %v, want empty", got)
	}
}

func TestReplaceCount(t *testing.T) {
	repl := map[string]string{
		"\u2705": "[PASS]",
		"\u274c": "[FAIL]",
	}

	tests := []struct {
		name         string
		input        string
		replacements map[string]string
		wantText     string
		wantCount    int
	}{
		{
			name:         "no emoji unchanged, count zero",
			input:        "plain text",
			replacements: repl,
			wantText:     "plain text",
			wantCount:    0,
		},
		{
			name:         "single substitution",
			input:        "\u2705 build",
			replacements: repl,
			wantText:     "[PASS] build",
			wantCount:    1,
		},
		{
			name:         "two substitutions",
			input:        "\u2705 pass \u274c fail",
			replacements: repl,
			wantText:     "[PASS] pass [FAIL] fail",
			wantCount:    2,
		},
		{
			name:         "unmapped emoji stripped and counted",
			input:        "\u2705 check \U0001F680 rocket",
			replacements: repl,
			wantText:     "[PASS] check  rocket",
			wantCount:    2,
		},
		{
			name:         "nil map behaves like Demojify",
			input:        "\U0001F680 launch",
			replacements: nil,
			wantText:     " launch",
			wantCount:    1,
		},
		{
			name:         "empty string",
			input:        "",
			replacements: repl,
			wantText:     "",
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotCount := demojify.ReplaceCount(tt.input, tt.replacements)
			if gotText != tt.wantText {
				t.Errorf("ReplaceCount(%q) text = %q, want %q", tt.input, gotText, tt.wantText)
			}
			if gotCount != tt.wantCount {
				t.Errorf("ReplaceCount(%q) count = %d, want %d", tt.input, gotCount, tt.wantCount)
			}
		})
	}
}

func TestFindMatchesInFile(t *testing.T) {
	repl := demojify.DefaultReplacements()

	t.Run("file with emoji returns matches with line and column", func(t *testing.T) {
		dir := t.TempDir()
		// Line 1: checkmark at column 0; line 2: cross mark at column 0
		path := writeTempFile(t, dir, "doc.md", "\u2705 passed\n\u274c failed\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if len(matches) != 2 {
			t.Fatalf("got %d matches, want 2", len(matches))
		}

		m0 := matches[0]
		if m0.Emoji != "\u2705" {
			t.Errorf("matches[0].Emoji = %q, want checkmark", m0.Emoji)
		}
		if m0.Replacement != "[PASS]" {
			t.Errorf("matches[0].Replacement = %q, want [PASS]", m0.Replacement)
		}
		if m0.Line != 1 {
			t.Errorf("matches[0].Line = %d, want 1", m0.Line)
		}
		if m0.Column != 0 {
			t.Errorf("matches[0].Column = %d, want 0", m0.Column)
		}
		if m0.Context == "" {
			t.Error("matches[0].Context should not be empty")
		}

		m1 := matches[1]
		if m1.Line != 2 {
			t.Errorf("matches[1].Line = %d, want 2", m1.Line)
		}
		if m1.Emoji != "\u274c" {
			t.Errorf("matches[1].Emoji = %q, want cross mark", m1.Emoji)
		}
		if m1.Replacement != "[FAIL]" {
			t.Errorf("matches[1].Replacement = %q, want [FAIL]", m1.Replacement)
		}
	})

	t.Run("file with no emoji returns nil", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "clean.txt", "This file has no emoji\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if matches != nil {
			t.Errorf("got %d matches, want nil for clean file", len(matches))
		}
	})

	t.Run("unmapped emoji has empty replacement", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTempFile(t, dir, "log.txt", "\U0001F680 deployed\n")

		matches, err := demojify.FindMatchesInFile(path, repl)
		if err != nil {
			t.Fatalf("FindMatchesInFile: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("expected at least one match for rocket emoji")
		}
		// Rocket is not in DefaultReplacements; replacement should be empty.
		if matches[0].Replacement != "" {
			t.Errorf("Replacement = %q, want empty for unmapped emoji", matches[0].Replacement)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := demojify.FindMatchesInFile("/nonexistent/no-file.txt", repl)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}
