package demojify

import (
	"regexp"
	"strings"
)

var (
	// crlfReplacer normalizes Windows (CRLF) and old Mac (CR) line endings to
	// Unix LF before any regex-based whitespace processing runs. CRLF is replaced
	// before bare CR so that a sequence like \r\n is not double-converted.
	crlfReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

	// trailingSpaceRE removes trailing spaces and tabs before a newline.
	trailingSpaceRE = regexp.MustCompile(`[ \t]+\n`)

	// multiNewlineRE collapses three or more consecutive newlines to two,
	// limiting blank lines to a single blank line between paragraphs.
	multiNewlineRE = regexp.MustCompile(`\n{3,}`)
)

// collapseInlineSpaces collapses runs of 2+ spaces/tabs to a single space,
// but only after the first non-whitespace character on each line. Leading
// indentation is left untouched.
func collapseInlineSpaces(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Find the first non-space, non-tab byte.
		firstNonWS := -1
		for j := 0; j < len(line); j++ {
			if line[j] != ' ' && line[j] != '\t' {
				firstNonWS = j
				break
			}
		}
		if firstNonWS < 0 {
			// Line is all whitespace; trailingSpaceRE handles it.
			continue
		}
		// Preserve leading indentation; collapse runs in the remainder.
		prefix := line[:firstNonWS]
		rest := line[firstNonWS:]
		var b strings.Builder
		b.Grow(len(rest))
		inRun := false
		for j := 0; j < len(rest); j++ {
			ch := rest[j]
			if ch == ' ' || ch == '\t' {
				if !inRun {
					b.WriteByte(' ')
					inRun = true
				}
			} else {
				b.WriteByte(ch)
				inRun = false
			}
		}
		lines[i] = prefix + b.String()
	}
	return strings.Join(lines, "\n")
}

// Normalize collapses redundant whitespace in text while preserving
// leading indentation on each line:
//   - CRLF (\r\n) and bare CR (\r) line endings are converted to LF (\n),
//   - consecutive horizontal spaces or tabs AFTER the first non-whitespace
//     character on a line are reduced to one space (indentation is kept),
//   - trailing whitespace before a newline is removed,
//   - three or more consecutive blank lines are collapsed to two.
//
// The returned string is trimmed of leading and trailing whitespace.
//
// Because leading indentation is preserved, Normalize is safe to use on
// Markdown files with nested lists and indented code blocks. However,
// inline runs of multiple spaces or tabs after the first non-whitespace
// character are collapsed to a single space, which breaks column-aligned
// comments and tabular formatting. Use a formatter such as gofmt to
// restore comment alignment in Go source files after normalizing.
func Normalize(text string) string {
	text = crlfReplacer.Replace(text)
	text = collapseInlineSpaces(text)
	text = trailingSpaceRE.ReplaceAllString(text, "\n")
	text = multiNewlineRE.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
