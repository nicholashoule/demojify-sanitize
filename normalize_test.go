package demojify_test

import (
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no changes needed",
			input: "Hello, World!",
			want:  "Hello, World!",
		},
		{
			name:  "multiple spaces collapsed",
			input: "Hello   World",
			want:  "Hello World",
		},
		{
			name:  "tabs collapsed",
			input: "Hello\t\tWorld",
			want:  "Hello World",
		},
		{
			name:  "trailing whitespace before newline removed",
			input: "Hello   \nWorld",
			want:  "Hello\nWorld",
		},
		{
			name:  "three blank lines collapsed to one",
			input: "Hello\n\n\n\nWorld",
			want:  "Hello\n\nWorld",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "  Hello World  ",
			want:  "Hello World",
		},
		{
			name:  "mixed redundant whitespace",
			input: "Hello   World  \n\n\nMore text",
			want:  "Hello World\n\nMore text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   \n\n\t  \n",
			want:  "",
		},
		{
			name:  "two blank lines unchanged",
			input: "Para one.\n\nPara two.",
			want:  "Para one.\n\nPara two.",
		},
		{
			name:  "CRLF line endings normalized to LF",
			input: "Hello\r\nWorld\r\n",
			want:  "Hello\nWorld",
		},
		{
			name:  "bare CR line endings normalized to LF",
			input: "Hello\rWorld\r",
			want:  "Hello\nWorld",
		},
		{
			name:  "CRLF with trailing spaces cleaned correctly",
			input: "Hello   \r\nWorld\r\n",
			want:  "Hello\nWorld",
		},
		{
			name:  "CRLF excess blank lines collapsed",
			input: "A\r\n\r\n\r\n\r\nB",
			want:  "A\n\nB",
		},
		{
			name:  "mixed CRLF and LF normalized",
			input: "line1\r\nline2\nline3\r\n",
			want:  "line1\nline2\nline3",
		},
		// --- Indentation preservation ---
		{
			name:  "Markdown nested list preserved",
			input: "- item\n  - nested\n    - deep",
			want:  "- item\n  - nested\n    - deep",
		},
		{
			name:  "Markdown indented code block preserved",
			input: "Paragraph\n\n    code line 1\n    code line 2",
			want:  "Paragraph\n\n    code line 1\n    code line 2",
		},
		{
			name:  "tab indentation preserved inline collapsed",
			input: "line1\n\t\tindented  text  here\nline3",
			want:  "line1\n\t\tindented text here\nline3",
		},
		{
			name:  "inline multi-space collapsed after content",
			input: "x := 1   // aligned",
			want:  "x := 1 // aligned",
		},
		{
			name:  "leading spaces preserved multiline",
			input: "func main() {\n\tfmt.Println(\"hi\")\n}",
			want:  "func main() {\n\tfmt.Println(\"hi\")\n}",
		},
		{
			name:  "mixed indent and inline collapse",
			input: "top\n  key:    value\n    nested:    data",
			want:  "top\n  key: value\n    nested: data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q)\n  got  %q\n  want %q", tt.input, got, tt.want)
			}
		})
	}
}
