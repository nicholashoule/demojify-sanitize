package demojify_test

import (
	"strings"
	"sync"
	"testing"
	"unicode"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDemojify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no emojis",
			input: "Hello, World!",
			want:  "Hello, World!",
		},
		{
			name:  "emoticons",
			input: "Hello 😀 World 🎉",
			want:  "Hello  World ",
		},
		{
			name:  "misc symbols check and cross",
			input: "Done ✅ Error ❌",
			want:  "Done  Error ",
		},
		{
			name:  "misc symbols star and sparkles",
			input: "Featured ⭐ New ✨",
			want:  "Featured  New ",
		},
		{
			name:  "regional indicators form a flag",
			input: "Flag: 🇺🇸",
			want:  "Flag: ",
		},
		{
			name:  "ZWJ family sequence",
			input: "Family: 👨\u200D👩\u200D👧",
			want:  "Family: ",
		},
		{
			name:  "variation selector 16",
			input: "Star: ⭐️",
			want:  "Star: ",
		},
		{
			name:  "dingbat scissors",
			input: "Cut: ✂ here",
			want:  "Cut:  here",
		},
		{
			name:  "markdown text is preserved",
			input: "# Heading\n\nSome **bold** text.",
			want:  "# Heading\n\nSome **bold** text.",
		},
		{
			name:  "non-emoji unicode is preserved",
			input: "中文 العربية Ñoño",
			want:  "中文 العربية Ñoño",
		},
		{
			name:  "information source symbol (U+2139)",
			input: "ℹ️ read the docs",
			want:  " read the docs",
		},
		{
			name: "subdivision flag England (U+1F3F4 + TAG sequence)",
			// 🏴󠁧󠁢󠁥󠁮󠁧󠁿 = U+1F3F4 U+E0067 U+E0062 U+E0065 U+E006E U+E0067 U+E007F
			input: "Location: \U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F",
			want:  "Location: ",
		},
		{
			name:  "mixed emoji and text",
			input: "🚀 Deploy complete! Check the dashboard 📊",
			want:  " Deploy complete! Check the dashboard ",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only emoji returns empty string",
			input: "\U0001F680\U0001F4CA\u2705",
			want:  "",
		},
		{
			name:  "skin tone modifier stripped",
			input: "wave \U0001F44B\U0001F3FD end",
			want:  "wave  end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Demojify(tt.input)
			if got != tt.want {
				t.Errorf("Demojify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"has emoji", "Hello 😀", true},
		{"no emoji", "Hello, World!", false},
		{"only emoji", "🎉", true},
		{"misc symbol", "Done ✅", true},
		{"non-emoji unicode", "中文", false},
		{"empty string", "", false},
		{"information source U+2139", "ℹ note", true},
		{"tag character U+E007F", "end\U000E007F", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.ContainsEmoji(tt.input)
			if got != tt.want {
				t.Errorf("ContainsEmoji(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCountEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 0},
		{"no emoji", "Hello, World!", 0},
		{"one emoji", "Hello \U0001F600", 1},
		{"two distinct emoji", "Hello \U0001F600 World \U0001F680", 2},
		{"repeated emoji", "\u2705 pass \u2705 pass \u2705", 3},
		{"ZWJ sequence counted per codepoint", "Family: \U0001F468\u200D\U0001F469\u200D\U0001F467", 5},
		{"variation selector counted", "\u26A0\uFE0F warning", 2},
		{"plain ASCII", "all good", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.CountEmoji(tt.input)
			if got != tt.want {
				t.Errorf("CountEmoji(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestBytesSaved(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 0},
		{"no emoji", "Hello, World!", 0},
		{"one 4-byte emoji", "Hello \U0001F600", 4},
		{"multiple emoji", "\u2705 pass \u274C fail", 6},
		{"emoji only", "\U0001F680\U0001F4CA", 8},
		{"plain text", "no emoji here", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.BytesSaved(tt.input)
			if got != tt.want {
				t.Errorf("BytesSaved(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTechnicalSymbolRanges(t *testing.T) {
	ranges := demojify.TechnicalSymbolRanges()
	if len(ranges) == 0 {
		t.Fatal("TechnicalSymbolRanges() returned empty slice")
	}

	// Symbols that should be in the ranges and are also matched by emojiRE.
	covered := []struct {
		name string
		r    rune
	}{
		{"check mark", '\u2713'},
		{"heavy check mark", '\u2714'},
		{"ballot x", '\u2717'},
		{"warning sign", '\u26A0'},
		{"gear", '\u2699'},
		{"black star", '\u2605'},
		{"card suit spade", '\u2660'},
	}
	for _, tt := range covered {
		t.Run(tt.name, func(t *testing.T) {
			if !unicode.IsOneOf(ranges, tt.r) {
				t.Errorf("TechnicalSymbolRanges should cover %s (U+%04X)", tt.name, tt.r)
			}
		})
	}

	// Verify that AllowedRanges with TechnicalSymbolRanges preserves these symbols.
	t.Run("preserves technical symbols during sanitization", func(t *testing.T) {
		input := "warning \u26A0 and gear \u2699 but rocket \U0001F680"
		opts := demojify.Options{
			RemoveEmojis:  true,
			AllowedRanges: demojify.TechnicalSymbolRanges(),
		}
		got := demojify.Sanitize(input, opts)
		want := "warning \u26A0 and gear \u2699 but rocket "
		if got != want {
			t.Errorf("Sanitize with TechnicalSymbolRanges:\ngot  %q\nwant %q", got, want)
		}
	})
}

// TestDemojifyKeycapSequences verifies that keycap emoji sequences
// (digit + VS16 + Combining Enclosing Keycap U+20E3) are handled
// correctly. The digit itself is plain ASCII and must be preserved;
// only the VS16 and keycap codepoints are stripped.
func TestDemojifyKeycapSequences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "keycap 1 sequence",
			input: "Press 1\uFE0F\u20E3 to continue",
			want:  "Press 1 to continue",
		},
		{
			name:  "keycap hash",
			input: "Dial #\uFE0F\u20E3 for help",
			want:  "Dial # for help",
		},
		{
			name:  "keycap asterisk",
			input: "Press *\uFE0F\u20E3 now",
			want:  "Press * now",
		},
		{
			name:  "multiple keycaps in sequence",
			input: "Code: 1\uFE0F\u20E3 2\uFE0F\u20E3 3\uFE0F\u20E3",
			want:  "Code: 1 2 3",
		},
		{
			name:  "bare combining enclosing keycap stripped",
			input: "x\u20E3 alone",
			want:  "x alone",
		},
		{
			name:  "ContainsEmoji detects keycap",
			input: "1\uFE0F\u20E3",
			want:  "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Demojify(tt.input)
			if got != tt.want {
				t.Errorf("Demojify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}

	// Verify ContainsEmoji returns true for keycap sequences.
	if !demojify.ContainsEmoji("1\uFE0F\u20E3") {
		t.Error("ContainsEmoji(keycap 1) = false, want true")
	}
}

// TestDemojifyVeryLongSingleLine verifies that Demojify handles a very long
// single line (no newlines) without excessive backtracking or performance
// degradation. This is a regression guard against pathological regex behavior.
func TestDemojifyVeryLongSingleLine(t *testing.T) {
	// Build a 1 MB string with emoji scattered throughout.
	const chunk = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. "
	var b strings.Builder
	for b.Len() < 1<<20 {
		b.WriteString(chunk)
		if b.Len()%500 < len(chunk) {
			b.WriteString("\U0001F680")
		}
	}
	input := b.String()

	got := demojify.Demojify(input)
	if demojify.ContainsEmoji(got) {
		t.Error("Demojify output still contains emoji on very long single line")
	}
	if len(got) >= len(input) {
		t.Error("Demojify output should be shorter than input (emoji removed)")
	}
	// Verify non-emoji content is intact.
	if !strings.Contains(got, "Lorem ipsum") {
		t.Error("Demojify removed non-emoji content")
	}
}

// TestDemojifyConcurrent verifies that Demojify and ContainsEmoji are safe
// for concurrent use from multiple goroutines, exercising the compiled
// package-level regex without data races.
func TestDemojifyConcurrent(_ *testing.T) {
	const goroutines = 50
	inputs := []string{
		"Hello \U0001F680 World",
		"No emoji here",
		"\u2705 pass \u274C fail",
		"Mix \U0001F600 of \U0001F4CA content \U0001F389",
		"",
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			input := inputs[idx%len(inputs)]
			_ = demojify.Demojify(input)
			_ = demojify.ContainsEmoji(input)
			_ = demojify.CountEmoji(input)
			_ = demojify.BytesSaved(input)
			_ = demojify.FindAll(input)
		}(i)
	}
	wg.Wait()
}

// TestSanitizeBuildPlaceholderCollisionFallback exercises the extreme fallback
// path in buildPlaceholders, which triggers when all 34 Unicode noncharacter
// sentinels (U+FDD0-U+FDEF, U+FFFE, U+FFFF) collide with text already in the
// input. The function must still return correct unique placeholders via the
// three-rune prefix fallback, allowing AllowedEmojis to work correctly.
func TestSanitizeBuildPlaceholderCollisionFallback(t *testing.T) {
	// Build text that contains the placeholder string for every possible
	// single-character sentinel, forcing all 34 candidates to collide.
	// Each sentinel (prefix) produces placeholder prefix+"0"+prefix for n=1.
	sentinels := []rune{
		0xFDD0, 0xFDD1, 0xFDD2, 0xFDD3, 0xFDD4, 0xFDD5, 0xFDD6, 0xFDD7,
		0xFDD8, 0xFDD9, 0xFDDA, 0xFDDB, 0xFDDC, 0xFDDD, 0xFDDE, 0xFDDF,
		0xFDE0, 0xFDE1, 0xFDE2, 0xFDE3, 0xFDE4, 0xFDE5, 0xFDE6, 0xFDE7,
		0xFDE8, 0xFDE9, 0xFDEA, 0xFDEB, 0xFDEC, 0xFDED, 0xFDEE, 0xFDEF,
		0xFFFE, 0xFFFF,
	}
	var collision strings.Builder
	for _, s := range sentinels {
		prefix := string(s)
		collision.WriteString(prefix + "0" + prefix)
	}

	// The allowed emoji must survive; the grinning face must be stripped.
	text := collision.String() + "\U0001F680 launch \U0001F600"
	opts := demojify.Options{
		RemoveEmojis:  true,
		AllowedEmojis: []string{"\U0001F680"},
	}
	result := demojify.Sanitize(text, opts)

	if !strings.Contains(result, "\U0001F680") {
		t.Errorf("AllowedEmojis did not preserve rocket in collision-fallback path: %q", result)
	}
	if strings.Contains(result, "\U0001F600") {
		t.Errorf("grinning face should be stripped in collision-fallback path: %q", result)
	}
	if !strings.Contains(result, "launch") {
		t.Errorf("plain text 'launch' is missing from result: %q", result)
	}
}

// TestDemojifyPreservesLegalSymbols verifies that ©, ®, and ™ are never
// stripped by Demojify. These codepoints (U+00A9, U+00AE, U+2122) have the
// Unicode emoji property in some contexts but are standard legal/documentation
// characters that must be preserved in source files.
func TestDemojifyPreservesLegalSymbols(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"copyright sign U+00A9", "Copyright \u00a9 2024 Example Corp."},
		{"registered sign U+00AE", "Registered trademark\u00ae"},
		{"trade mark sign U+2122", "Example\u2122 is a trademark."},
		{"all three together", "\u00a9 \u00ae \u2122 in one line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Demojify(tt.input)
			if got != tt.input {
				t.Errorf("Demojify(%q) = %q; want input unchanged", tt.input, got)
			}
		})
	}
}
