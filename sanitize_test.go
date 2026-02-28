package demojify_test

import (
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDefaultOptions(t *testing.T) {
	opts := demojify.DefaultOptions()
	if !opts.RemoveEmojis {
		t.Error("DefaultOptions().RemoveEmojis should be true")
	}
	if !opts.RemoveAIClutter {
		t.Error("DefaultOptions().RemoveAIClutter should be true")
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
			name:  "remove AI clutter only – phrase alone on line",
			input: "Certainly!\nHello World",
			opts:  demojify.Options{RemoveAIClutter: true},
			want:  "Hello World",
		},
		{
			name:  "remove AI clutter only – phrase followed by content",
			input: "Sure! Here is the answer.",
			opts:  demojify.Options{RemoveAIClutter: true},
			want:  "Here is the answer.",
		},
		{
			name:  "remove AI clutter only – phrase not at line start is preserved",
			input: "He said: Sure! Here is the answer.",
			opts:  demojify.Options{RemoveAIClutter: true},
			want:  "He said: Sure! Here is the answer.",
		},
		{
			name:  "remove AI clutter only – no false positive on 'Sure enough'",
			input: "Sure enough, the build passed.",
			opts:  demojify.Options{RemoveAIClutter: true},
			want:  "Sure enough, the build passed.",
		},
		{
			name:  "normalize whitespace only",
			input: "Hello  World\n\n\nMore text",
			opts:  demojify.Options{NormalizeWhitespace: true},
			want:  "Hello World\n\nMore text",
		},
		{
			name:  "all options – emoji on next line",
			input: "Sure!\n🚀 Deploy complete!\n\n\nCheck the dashboard 📊",
			opts:  demojify.DefaultOptions(),
			want:  "Deploy complete!\n\nCheck the dashboard",
		},
		{
			name:  "all options – Of course variant",
			input: "Of course! Here is the answer.\n\n\nDone.",
			opts:  demojify.DefaultOptions(),
			want:  "Here is the answer.\n\nDone.",
		},
		{
			name:  "all options – I'd be happy to",
			input: "I'd be happy to help!\nHere is the code.",
			opts:  demojify.DefaultOptions(),
			want:  "Here is the code.",
		},
		{
			name:  "all options – I hope this helps",
			input: "The answer is 42.\nI hope this helps!",
			opts:  demojify.DefaultOptions(),
			want:  "The answer is 42.",
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
