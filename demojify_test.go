package demojify_test

import (
	"testing"

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
