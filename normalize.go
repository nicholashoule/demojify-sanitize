package demojify

import (
	"regexp"
	"strings"
)

// crlfReplacer normalizes Windows (CRLF) and old Mac (CR) line endings to
// Unix LF before any regex-based whitespace processing runs. CRLF is replaced
// before bare CR so that a sequence like \r\n is not double-converted.
var crlfReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

var (
	// multiSpaceRE collapses consecutive horizontal whitespace (tabs and
	// spaces, but not newlines) to a single space.
	multiSpaceRE = regexp.MustCompile(`[^\S\n]+`)

	// trailingSpaceRE removes trailing spaces and tabs before a newline.
	trailingSpaceRE = regexp.MustCompile(`[ \t]+\n`)

	// multiNewlineRE collapses three or more consecutive newlines to two,
	// limiting blank lines to a single blank line between paragraphs.
	multiNewlineRE = regexp.MustCompile(`\n{3,}`)
)

// Normalize collapses redundant whitespace in text:
//   - CRLF (\r\n) and bare CR (\r) line endings are converted to LF (\n),
//   - consecutive horizontal spaces or tabs are reduced to one space,
//   - trailing whitespace before a newline is removed,
//   - three or more consecutive blank lines are collapsed to two.
//
// The returned string is trimmed of leading and trailing whitespace.
func Normalize(text string) string {
	text = crlfReplacer.Replace(text)
	text = multiSpaceRE.ReplaceAllString(text, " ")
	text = trailingSpaceRE.ReplaceAllString(text, "\n")
	text = multiNewlineRE.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
