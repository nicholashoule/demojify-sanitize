package demojify_test

import (
	"strings"
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
		{
			name:         "empty key in map is silently ignored",
			input:        "hello world",
			replacements: map[string]string{"": "INSERTED", "\u2705": "[PASS]"},
			want:         "hello world",
		},
		{
			name:         "adjacent emoji produce single token (space-separated)",
			input:        "\u26a0 \u26a0 important",
			replacements: map[string]string{"\u26a0": "WARNING"},
			want:         "WARNING important",
		},
		{
			name:         "adjacent emoji produce single token (concatenated)",
			input:        "\u26a0\u26a0 important",
			replacements: map[string]string{"\u26a0": "WARNING"},
			want:         "WARNING important",
		},
		{
			name:         "three adjacent emoji collapsed to one token",
			input:        "\u26a0 \u26a0 \u26a0 triple",
			replacements: map[string]string{"\u26a0": "WARNING"},
			want:         "WARNING triple",
		},
		{
			name:         "non-repeated tokens are not affected",
			input:        "\u2705 test \u274c fail",
			replacements: map[string]string{"\u2705": "PASS", "\u274c": "FAIL"},
			want:         "PASS test FAIL fail",
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

func TestFindAllMappedEmptyKey(t *testing.T) {
	repl := map[string]string{"": "INSERTED", "\u2705": "[PASS]"}
	got := demojify.FindAllMapped("hello \u2705", repl)
	// Empty key must be silently skipped; only the check mark is found.
	if len(got) != 1 || got[0] != "\u2705" {
		t.Errorf("FindAllMapped with empty key = %v, want [\u2705]", got)
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
		{
			name:         "empty key in map is silently ignored",
			input:        "hello world",
			replacements: map[string]string{"": "INSERTED"},
			wantText:     "hello world",
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

// TestReplaceDefaultReplacementsPreservesASCII is a regression test for the
// bug reported by a downstream consumer where running Replace with
// DefaultReplacements() destroyed plain ASCII content: "//" was collapsed to
// "/", "**" to "*", "--" to "-", and "oo" in words like "root" and "bool" to
// "o". The root cause was collapseRepeatedTokens running against single-
// character replacement values (e.g. "/" from U+2797, "*" from U+2022, "-"
// from U+2796, "o" from U+25CB) that appear legitimately throughout source
// code and documentation.
func TestReplaceDefaultReplacementsPreservesASCII(t *testing.T) {
	repl := demojify.DefaultReplacements()

	tests := []struct {
		name         string
		input        string
		wantContains string // non-empty: output must contain this substring (emoji was transformed)
		wantChanged  bool   // true: output must differ from input
	}{
		{name: "URL double slash preserved", input: "https://example.com/path"},
		{name: "Go comment double slash preserved", input: "// this is a comment"},
		{name: "Markdown bold preserved", input: "**bold text**"},
		{name: "CLI double dash preserved", input: "--flag value"},
		{name: "triple dash separator preserved", input: "---"},
		{name: "word with oo preserved", input: "root bool tool"},
		{name: "word with oo preserved (capitalized)", input: "Roo Code"},
		{name: "double plus preserved", input: "count++"},
		{name: "double caret preserved", input: "x^^2"},
		{name: "full source-like line preserved", input: "if err != nil { // handle error\n\treturn nil, fmt.Errorf(\"err: %w\", err)\n}"},
		{name: "URL in markdown preserved", input: "See [docs](https://example.com/page#section) for details."},
		{
			name:         "mixed content emoji and ASCII",
			input:        "\u26a0 WARNING: see https://example.com for details -- read carefully",
			wantContains: "WARNING: see https://example.com for details -- read carefully",
			wantChanged:  true,
		},
	}

	// The inputs above contain no emoji (or only emoji that map to long tokens),
	// so the only ASCII sequences present must emerge from the output unchanged.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Replace(tt.input, repl)
			if tt.wantContains == "" && !tt.wantChanged {
				// Pure-ASCII inputs must be preserved exactly.
				if got != tt.input {
					t.Errorf("Replace(%q) = %q, want %q", tt.input, got, tt.input)
				}
				return
			}
			// Mixed content: ASCII surrounding the emoji must be preserved, and
			// the emoji itself must be transformed.
			// ⚠ (U+26A0) maps to "[WARNING]"; "WARNING:" in the original text is
			// preserved unchanged since it is a different string from the token.
			if tt.wantContains != "" && !strings.Contains(got, tt.wantContains) {
				t.Errorf("Replace(%q) = %q, expected ASCII content %q to be preserved",
					tt.input, got, tt.wantContains)
			}
			if tt.wantChanged && got == tt.input {
				t.Errorf("Replace(%q) = %q, expected emoji to be transformed", tt.input, got)
			}
		})
	}

	// Explicit invariant checks: none of these double-char sequences must be collapsed.
	invariants := []struct {
		name        string
		input       string
		mustContain string
	}{
		{"double slash in URL", "visit https://go.dev/", "//"},
		{"double slash comment", "// TODO: fix", "//"},
		{"double star bold", "**important**", "**"},
		{"double dash flag", "run with --verbose", "--"},
		{"oo in root", "root directory", "oo"},
		{"oo in bool", "bool value", "oo"},
		{"oo in tool", "go tool", "oo"},
		{"double plus", "i++", "++"},
	}
	for _, tc := range invariants {
		t.Run(tc.name, func(t *testing.T) {
			got := demojify.Replace(tc.input, repl)
			if !strings.Contains(got, tc.mustContain) {
				t.Errorf("Replace(%q) = %q: ASCII sequence %q was incorrectly collapsed",
					tc.input, got, tc.mustContain)
			}
		})
	}
}

// TestReplaceLongTokensStillCollapse verifies that the fix to
// collapseRepeatedTokens did not break the intended behavior: adjacent
// identical emoji that both map to a long label token are still collapsed to a
// single occurrence.
func TestReplaceLongTokensStillCollapse(t *testing.T) {
	repl := demojify.DefaultReplacements()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "adjacent warning emoji collapse to one [WARNING]",
			input: "\u26a0 \u26a0",
			want:  "[WARNING]",
		},
		{
			name:  "concatenated warning emoji collapse to one [WARNING]",
			input: "\u26a0\u26a0",
			want:  "[WARNING]",
		},
		{
			name:  "adjacent fail emoji collapse to one [FAIL]",
			input: "\u274c \u274c build",
			want:  "[FAIL] build",
		},
		{
			name:  "three pass emoji collapse to one [PASS]",
			input: "\u2705 \u2705 \u2705 ok",
			want:  "[PASS] ok",
		},
		{
			name:  "adjacent deploy emoji collapse to one [DEPLOY]",
			input: "\U0001f680 \U0001f680",
			want:  "[DEPLOY]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Replace(tt.input, repl)
			if got != tt.want {
				t.Errorf("Replace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestReplaceConcurrent verifies that Replace and ReplaceCount are safe for
// concurrent use from multiple goroutines sharing the same replacements map
// (read-only). The race detector (go test -race) will catch any data races.
