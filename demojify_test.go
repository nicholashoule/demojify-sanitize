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
			input: "Family: 👨‍👩‍👧",
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
			name:  "mixed emoji and text",
			input: "🚀 Deploy complete! Check the dashboard 📊",
			want:  " Deploy complete! Check the dashboard ",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
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
