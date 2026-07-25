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

// collapseLineSpaces collapses runs of 2+ spaces/tabs in a single line to one
// space, but only after the first non-whitespace character. Leading
// indentation is left untouched. All-whitespace lines are returned unchanged;
// callers handle their trailing whitespace separately.
func collapseLineSpaces(line string) string {
	// Find the first non-space, non-tab byte.
	firstNonWS := -1
	for j := 0; j < len(line); j++ {
		if line[j] != ' ' && line[j] != '\t' {
			firstNonWS = j
			break
		}
	}
	if firstNonWS < 0 {
		return line
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
	return prefix + b.String()
}

// collapseInlineSpaces applies [collapseLineSpaces] to every line of text.
// All-whitespace lines are left for trailingSpaceRE to handle.
func collapseInlineSpaces(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = collapseLineSpaces(line)
	}
	return strings.Join(lines, "\n")
}

// tidyLine collapses inline whitespace runs in line and trims its trailing
// spaces and tabs. It is the per-line cleanup applied to lines that emoji
// removal or substitution actually changed.
func tidyLine(line string) string {
	return strings.TrimRight(collapseLineSpaces(line), " \t")
}

// tidyChangedLines applies [tidyLine] only to the lines of cleaned that
// differ from the corresponding line of original, leaving every untouched
// line byte-for-byte identical. This keeps whitespace artifacts of emoji
// removal from surviving a fix pass while preserving column-aligned comments
// and tab alignment on lines the removal never touched.
//
// Both inputs must use bare-LF line endings. When the two inputs have a
// different line count -- possible only when a replacement value inserts or
// removes newlines -- per-line pairing is meaningless and every line of
// cleaned is tidied instead.
func tidyChangedLines(original, cleaned string) string {
	origLines := strings.Split(original, "\n")
	lines := strings.Split(cleaned, "\n")
	if len(origLines) != len(lines) {
		for i := range lines {
			lines[i] = tidyLine(lines[i])
		}
		return strings.Join(lines, "\n")
	}
	for i := range lines {
		if lines[i] != origLines[i] {
			lines[i] = tidyLine(lines[i])
		}
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
