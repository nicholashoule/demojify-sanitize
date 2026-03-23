package demojify_test

import (
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDefaultReplacements(t *testing.T) {
	t.Run("returns non-nil non-empty map", func(t *testing.T) {
		m := demojify.DefaultReplacements()
		if m == nil {
			t.Fatal("DefaultReplacements() returned nil")
		}
		if len(m) == 0 {
			t.Fatal("DefaultReplacements() returned empty map")
		}
	})

	t.Run("returns independent copy each call", func(t *testing.T) {
		m1 := demojify.DefaultReplacements()
		m2 := demojify.DefaultReplacements()
		m1["\u2705"] = "MUTATED"
		if m2["\u2705"] == "MUTATED" {
			t.Error("mutation of one copy affected another -- not independent copies")
		}
	})
}

// TestDefaultReplacementsEntries verifies that key entries across all
// six categories are present with correct text equivalents.
func TestDefaultReplacementsEntries(t *testing.T) {
	m := demojify.DefaultReplacements()

	tests := []struct {
		category string
		key      string
		want     string
	}{
		// Warning and alerts
		{"warning bare", "\u26a0", "[WARNING]"},
		{"warning with selector", "\u26a0\ufe0f", "[WARNING]"},
		{"double exclamation", "\u203c", "[ALERT]"},

		// Status symbols
		{"checkmark", "\u2705", "[PASS]"},
		{"heavy check", "\u2714", "[PASS]"},
		{"heavy check with selector", "\u2714\ufe0f", "[PASS]"},
		{"cross mark", "\u274c", "[FAIL]"},
		{"heavy ballot X", "\u2718", "[FAIL]"},
		{"exclamation", "\u2757", "[ALERT]"},

		// Favorites and highlights
		{"star", "\u2b50", "[FEATURED]"},
		{"filled star", "\u2605", "[FEATURED]"},
		{"light bulb", "\U0001f4a1", "[TIP]"},
		{"pushpin", "\U0001f4cc", "[PINNED]"},
		{"key", "\U0001f511", "[KEY]"},
		{"locked", "\U0001f512", "[LOCKED]"},
		{"unlocked", "\U0001f513", "[UNLOCKED]"},

		// Information symbol
		{"info bare", "\u2139", "[INFO]"},
		{"info with selector", "\u2139\ufe0f", "[INFO]"},

		// Severity indicators (colored circles)
		{"red circle - error", "\U0001f534", "[ERROR]"},
		{"orange circle - warning", "\U0001f7e0", "[WARNING]"},
		{"yellow circle - caution", "\U0001f7e1", "[CAUTION]"},
		{"green circle - ok", "\U0001f7e2", "[OK]"},
		{"blue circle - info", "\U0001f535", "[INFO]"},
		{"black circle - inactive", "\u26ab", "[INACTIVE]"},
		{"white circle - inactive", "\u26aa", "[INACTIVE]"},

		// Stop and prohibition
		{"stop sign", "\U0001f6d1", "[STOP]"},
		{"no entry", "\u26d4", "[NO ENTRY]"},
		{"prohibited", "\U0001f6ab", "[PROHIBITED]"},

		// Cloud and deployment
		{"cloud", "\u2601", "[CLOUD]"},
		{"cloud with selector", "\u2601\ufe0f", "[CLOUD]"},
		{"chart increasing", "\U0001f4c8", "[GROWTH]"},
		{"notebook", "\U0001f4da", "[DOCS]"},
		{"gear", "\u2699", "[CONFIG]"},
		{"gear with selector", "\u2699\ufe0f", "[CONFIG]"},
		{"laptop", "\U0001f4bb", "[CODE]"},

		// CI/CD workflow
		{"rocket - deploy", "\U0001f680", "[DEPLOY]"},
		{"package", "\U0001f4e6", "[PACKAGE]"},
		{"party popper - success", "\U0001f389", "[SUCCESS]"},
		{"sparkles - new", "\u2728", "[NEW]"},
		{"checkered flag - done", "\U0001f3c1", "[DONE]"},
		{"wrench - fix", "\U0001f527", "[FIX]"},
		{"tools", "\U0001f6e0", "[TOOLS]"},
		{"recycle bare", "\u267b", "[RECYCLE]"},
		{"recycle with selector", "\u267b\ufe0f", "[RECYCLE]"},
		{"floppy disk - save", "\U0001f4be", "[SAVE]"},
		{"fire - hot", "\U0001f525", "[HOT]"},
		{"hundred points", "\U0001f4af", "[100]"},

		// Math operators
		{"heavy multiplication x", "\u2716", "x"},
		{"heavy multiplication x with selector", "\u2716\ufe0f", "x"},
		{"heavy plus", "\u2795", "+"},
		{"heavy minus", "\u2796", "-"},
		{"heavy division", "\u2797", "/"},
		{"infinity bare", "\u267e", "[INFINITY]"},
		{"infinity with selector", "\u267e\ufe0f", "[INFINITY]"},

		// Arrows
		{"rightwards arrow", "\u2192", "->"},
		{"leftwards arrow", "\u2190", "<-"},
		{"upwards arrow", "\u2191", "^"},
		{"downwards arrow", "\u2193", "v"},
		{"double right arrow", "\u21d2", "=>"},
		{"black right with selector", "\u27a1\ufe0f", "->"},
		{"black left with selector", "\u2b05\ufe0f", "<-"},

		// Geometric shapes
		{"black circle", "\u25cf", "*"},
		{"white circle", "\u25cb", "o"},
		{"black square", "\u25a0", "*"},
		{"white square", "\u25a1", "[]"},
		{"black up triangle", "\u25b2", "^"},
		{"black down triangle", "\u25bc", "v"},
		{"black diamond", "\u25c6", "*"},
		{"black small square", "\u25aa", "*"},
		{"white small square", "\u25ab", "[]"},

		// Checkboxes
		{"ballot box", "\u2610", "[ ]"},
		{"ballot box check", "\u2611", "[x]"},
		{"ballot box X", "\u2612", "[x]"},

		// Dingbats
		{"bullet", "\u2022", "*"},
		{"triangular bullet", "\u2023", ">"},
		{"heart", "\u2764", "<3"},
		{"heart with selector", "\u2764\ufe0f", "<3"},
		{"diamond suit", "\u2666", "<>"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got, ok := m[tt.key]
			if !ok {
				t.Errorf("key %q not found in DefaultReplacements()", tt.key)
				return
			}
			if got != tt.want {
				t.Errorf("DefaultReplacements()[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestReplaceWithDefaultReplacements verifies end-to-end substitution using
// the built-in map across the main categories.
func TestReplaceWithDefaultReplacements(t *testing.T) {
	repl := demojify.DefaultReplacements()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "pass and fail markers",
			input: "\u2705 build \u274c deploy",
			want:  "[PASS] build [FAIL] deploy",
		},
		{
			name:  "warning bare codepoint",
			input: "\u26a0 check config",
			want:  "[WARNING] check config",
		},
		{
			name:  "warning variation selector preferred over bare",
			input: "\u26a0\ufe0f check config",
			want:  "[WARNING] check config",
		},
		{
			name:  "arrow substitution",
			input: "input \u2192 output",
			want:  "input -> output",
		},
		{
			name:  "bullet point",
			input: "\u2022 item",
			want:  "* item",
		},
		{
			name:  "ballot box unchecked",
			input: "\u2610 task",
			want:  "[ ] task",
		},
		{
			name:  "ballot box checked",
			input: "\u2611 done",
			want:  "[x] done",
		},
		{
			name:  "gear configuration",
			input: "\u2699\ufe0f settings",
			want:  "[CONFIG] settings",
		},
		{
			name:  "unmapped emoji is stripped",
			input: "\U0001F600 laugh", // grinning face, not in DefaultReplacements
			want:  " laugh",
		},
		{
			name:  "mixed: mapped and unmapped",
			input: "\u2705 tests \U0001F600 deployed",
			want:  "[PASS] tests  deployed",
		},
		{
			name:  "rocket deploy",
			input: "\U0001F680 launch to production",
			want:  "[DEPLOY] launch to production",
		},
		{
			name:  "severity colored circles",
			input: "\U0001f534 error \U0001f7e2 ok \U0001f7e1 caution",
			want:  "[ERROR] error [OK] ok [CAUTION] caution",
		},
		{
			name:  "information symbol",
			input: "\u2139\ufe0f read the docs",
			want:  "[INFO] read the docs",
		},
		{
			name:  "math operators",
			input: "a \u2795 b \u2796 c \u2716 d",
			want:  "a + b - c x d",
		},
		{
			name:  "no emoji unchanged",
			input: "plain text only",
			want:  "plain text only",
		},
		{
			name:  "tip lightbulb",
			input: "\U0001f4a1 use environment variables",
			want:  "[TIP] use environment variables",
		},
		{
			name:  "double right arrow",
			input: "a \u21d2 b",
			want:  "a => b",
		},
		{
			name:  "heart dingbat",
			input: "love \u2764\ufe0f code",
			want:  "love <3 code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Replace(tt.input, repl)
			if got != tt.want {
				t.Errorf("Replace(%q, DefaultReplacements()) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
