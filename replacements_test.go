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
		{"pointing right - see", "\U0001f449", "[SEE]"},

		// CI/CD workflow
		{"rocket - deploy", "\U0001f680", "[DEPLOY]"},
		{"package", "\U0001f4e6", "[PACKAGE]"},
		{"party popper - success", "\U0001f389", "[SUCCESS]"},
		{"sparkles - new", "\u2728", "[NEW]"},
		{"checkered flag - done", "\U0001f3c1", "[DONE]"},
		{"wrench - fix", "\U0001f527", "[FIX]"},
		{"tools", "\U0001f6e0", "[TOOLS]"},
		{"toolbox - tools", "\U0001f9f0", "[TOOLS]"},
		{"recycle bare", "\u267b", "[RECYCLE]"},
		{"recycle with selector", "\u267b\ufe0f", "[RECYCLE]"},
		{"floppy disk - save", "\U0001f4be", "[SAVE]"},
		{"fire - hot", "\U0001f525", "[HOT]"},
		{"hundred points", "\U0001f4af", "[100]"},
		{"rotating light - alert", "\U0001f6a8", "[ALERT]"},
		{"adhesive bandage - patch", "\U0001fa79", "[PATCH]"},

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

		// Heart variants
		{"blue heart", "\U0001f499", "[HEART]"},
		{"green heart", "\U0001f49a", "[HEART]"},
		{"yellow heart", "\U0001f49b", "[HEART]"},
		{"purple heart", "\U0001f49c", "[HEART]"},
		{"black heart", "\U0001f5a4", "[HEART]"},
		{"white heart", "\U0001f90d", "[HEART]"},
		{"brown heart", "\U0001f90e", "[HEART]"},
		{"orange heart", "\U0001f9e1", "[HEART]"},
		{"broken heart", "\U0001f494", "[HEART]"},

		// Project and issue tracking
		{"bug", "\U0001f41b", "[BUG]"},
		{"lady beetle", "\U0001f41e", "[BUG]"},
		{"breaking change", "\U0001f4a5", "[BREAKING]"},
		{"wip construction", "\U0001f6a7", "[CONSTRUCTION]"},
		{"test tube", "\U0001f9ea", "[TEST]"},
		{"bookmark release", "\U0001f516", "[RELEASE]"},
		{"label tag", "\U0001f3f7", "[TAG]"},
		{"label tag with selector", "\U0001f3f7\ufe0f", "[TAG]"},
		{"broom cleanup", "\U0001f9f9", "[CLEANUP]"},
		{"link", "\U0001f517", "[LINK]"},
		{"speech balloon comment", "\U0001f4ac", "[COMMENT]"},
		{"megaphone announce", "\U0001f4e3", "[ANNOUNCE]"},
		{"thumbs up approved", "\U0001f44d", "[APPROVED]"},
		{"thumbs down rejected", "\U0001f44e", "[REJECTED]"},
		{"puzzle plugin", "\U0001f9e9", "[PLUGIN]"},
		{"trophy award", "\U0001f3c6", "[AWARD]"},
		{"clipboard", "\U0001f4cb", "[CLIPBOARD]"},
		{"wastebasket trash", "\U0001f5d1", "[TRASH]"},
		{"wastebasket trash with selector", "\U0001f5d1\ufe0f", "[TRASH]"},
		{"gift", "\U0001f381", "[GIFT]"},
		{"gem", "\U0001f48e", "[GEM]"},

		// Colored squares
		{"red square error", "\U0001f7e5", "[ERROR]"},
		{"green square ok", "\U0001f7e9", "[OK]"},
		{"yellow square caution", "\U0001f7e8", "[CAUTION]"},
		{"blue square info", "\U0001f7e6", "[INFO]"},
		{"orange square warning", "\U0001f7e7", "[WARNING]"},
		{"black large square inactive", "\u2b1b", "[INACTIVE]"},
		{"white large square inactive", "\u2b1c", "[INACTIVE]"},

		// Media controls
		{"pause button", "\u23f8", "[PAUSED]"},
		{"pause button with selector", "\u23f8\ufe0f", "[PAUSED]"},
		{"stop button", "\u23f9", "[STOPPED]"},
		{"stop button with selector", "\u23f9\ufe0f", "[STOPPED]"},
		{"record button", "\u23fa", "[RECORDING]"},
		{"record button with selector", "\u23fa\ufe0f", "[RECORDING]"},
		{"fast-forward next", "\u23e9", "[NEXT]"},
		{"fast-rewind prev", "\u23ea", "[PREV]"},

		// Community and contributors
		{"ambulance hotfix", "\U0001f691", "[HOTFIX]"},
		{"twisted arrows merge", "\U0001f500", "[MERGE]"},
		{"clockwise arrows retry", "\U0001f501", "[RETRY]"},
		{"double up upgrade", "\u23eb", "[UPGRADE]"},
		{"double down downgrade", "\u23ec", "[DOWNGRADE]"},
		{"shield protected", "\U0001f6e1", "[PROTECTED]"},
		{"shield protected with selector", "\U0001f6e1\ufe0f", "[PROTECTED]"},
		{"robot bot", "\U0001f916", "[BOT]"},
		{"handshake contrib", "\U0001f91d", "[CONTRIB]"},
		{"bust user", "\U0001f464", "[USER]"},
		{"busts users", "\U0001f465", "[USERS]"},
		{"folded hands thanks", "\U0001f64f", "[THANKS]"},
		{"page facing up file", "\U0001f4c4", "[FILE]"},
		{"page with curl file", "\U0001f4c3", "[FILE]"},
		{"envelope email", "\U0001f4e7", "[EMAIL]"},
		{"money bag sponsor", "\U0001f4b0", "[SPONSOR]"},
		{"dollar banknote sponsor", "\U0001f4b5", "[SPONSOR]"},
		{"globe europe global", "\U0001f30d", "[GLOBAL]"},
		{"globe asia global", "\U0001f30f", "[GLOBAL]"},
		{"leftwards hook back", "\u21a9", "[BACK]"},
		{"leftwards hook back with selector", "\u21a9\ufe0f", "[BACK]"},
		{"rightwards hook forward", "\u21aa", "[FORWARD]"},
		{"rightwards hook forward with selector", "\u21aa\ufe0f", "[FORWARD]"},
		{"speaker mute", "\U0001f507", "[MUTE]"},
		{"bell mute", "\U0001f515", "[MUTE]"},

		// Media controls -- v0.8.0 additions (skip/prev track)
		{"next track skip", "\u23ed", "[SKIP]"},
		{"next track skip with selector", "\u23ed\ufe0f", "[SKIP]"},
		{"prev track", "\u23ee", "[PREV]"},
		{"prev track with selector", "\u23ee\ufe0f", "[PREV]"},

		// Community/status -- v0.8.0 additions
		{"small red triangle up", "\U0001f53c", "[UP]"},
		{"small red triangle down", "\U0001f53d", "[DOWN]"},
		{"backhand index up see", "\U0001f446", "[SEE]"},
		{"backhand index down see", "\U0001f447", "[SEE]"},
		{"backhand index left see", "\U0001f448", "[SEE]"},
		{"horizontal traffic light status", "\U0001f6a5", "[STATUS]"},
		{"vertical traffic light status", "\U0001f6a6", "[STATUS]"},

		// Platform and language indicators
		{"spouting whale docker", "\U0001f433", "[DOCKER]"},
		{"whale docker", "\U0001f40b", "[DOCKER]"},
		{"penguin linux", "\U0001f427", "[LINUX]"},
		{"snake python", "\U0001f40d", "[PYTHON]"},
		{"crab rust", "\U0001f980", "[RUST]"},
		{"hamster go", "\U0001f439", "[GO]"},
		{"red apple macos", "\U0001f34e", "[MACOS]"},
		{"window windows", "\U0001fa9f", "[WINDOWS]"},

		// Calendar and date indicators -- v0.8.0
		{"calendar date", "\U0001f4c5", "[DATE]"},
		{"tear-off calendar date", "\U0001f4c6", "[DATE]"},
		{"spiral calendar bare", "\U0001f5d3", "[CALENDAR]"},
		{"spiral calendar with selector", "\U0001f5d3\ufe0f", "[CALENDAR]"},

		// Scissors / removed -- v0.8.0
		{"scissors bare", "\u2702", "[REMOVED]"},
		{"scissors with selector", "\u2702\ufe0f", "[REMOVED]"},

		// Deprecated / tombstone -- v0.8.0
		{"headstone deprecated", "\U0001faa6", "[DEPRECATED]"},
		{"name badge deprecated", "\U0001f4db", "[DEPRECATED]"},

		// Flags -- v0.8.0: single-codepoint, ZWJ, tag-sequence, regional-indicator
		{"red flag", "\U0001f6a9", "[FLAG]"},
		{"white flag bare", "\U0001f3f3", "[FLAG]"},
		{"white flag with selector", "\U0001f3f3\ufe0f", "[FLAG]"},
		{"black flag", "\U0001f3f4", "[FLAG]"},
		{"crossed flags", "\U0001f38c", "[FLAG]"},
		{"rainbow flag ZWJ", "\U0001f3f3\ufe0f\u200d\U0001f308", "[FLAG]"},
		{"trans flag ZWJ", "\U0001f3f3\ufe0f\u200d\u26a7\ufe0f", "[FLAG]"},
		{"England subdivision flag", "\U0001f3f4\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f", "[FLAG]"},
		{"Scotland subdivision flag", "\U0001f3f4\U000e0067\U000e0062\U000e0073\U000e0063\U000e0074\U000e007f", "[FLAG]"},
		{"Wales subdivision flag", "\U0001f3f4\U000e0067\U000e0062\U000e0077\U000e006c\U000e0073\U000e007f", "[FLAG]"},
		{"US regional indicator", "\U0001f1fa\U0001f1f8", "[FLAG]"},
		{"GB regional indicator", "\U0001f1ec\U0001f1e7", "[FLAG]"},
		{"DE regional indicator", "\U0001f1e9\U0001f1ea", "[FLAG]"},
		{"ZA regional indicator", "\U0001f1ff\U0001f1e6", "[FLAG]"},
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
