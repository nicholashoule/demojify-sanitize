package demojify_test

import (
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
