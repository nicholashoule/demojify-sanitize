// Script-aware emoji removal: Options.ScriptAware and ContainsEmojiWith.
//
// emojiRE matches one codepoint at a time, so by default the Zero Width
// Joiner (U+200D) and the variation selectors (U+FE00-U+FE0F) are removed
// wherever they appear. Both also belong to ordinary text: Devanagari,
// Malayalam, Sinhala and other Indic scripts join letters with ZWJ, and
// CJK, Mongolian and mathematical text use standardized variation
// sequences (U+FE00-U+FE0D). ScriptAware removes those codepoints only
// where they are part of an emoji sequence. It is opt-in; the default
// behavior of every function is unchanged.
package demojify

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Codepoints the script-aware rules name. Written as numbers so no editor
// or sanitizer can strip an invisible literal from the source.
const (
	zwj                 = rune(0x200D) // Zero Width Joiner
	variationFirst      = rune(0xFE00) // first variation selector
	variationLastScript = rune(0xFE0D) // FE00-FE0D: standardized variation sequences
	presentationEmoji   = rune(0xFE0F) // VS16: emoji presentation (FE0E is text)
	combiningKeycap     = rune(0x20E3) // Combining Enclosing Keycap
)

// ContainsEmojiWith reports whether [Sanitize] with emoji removal and
// opts would remove at least one codepoint: the options-aware form of
// [ContainsEmoji]. It honors opts.ScriptAware, opts.AllowedRanges and
// opts.AllowedEmojis; opts.RemoveEmojis and opts.NormalizeWhitespace are
// ignored, since the question is always about emoji. With a zero Options
// it answers exactly as [ContainsEmoji].
//
// ContainsEmojiWith is safe for concurrent use.
func ContainsEmojiWith(text string, opts Options) bool {
	// The default set is a superset of what any option removes, so text
	// it finds clean is clean under every option.
	if !emojiRE.MatchString(text) {
		return false
	}
	if !opts.ScriptAware && len(opts.AllowedRanges) == 0 && !hasNonEmpty(opts.AllowedEmojis) {
		return true
	}
	return removeEmoji(text, opts) != text
}

// removeEmoji is the emoji-removal step of [Sanitize]. Without ScriptAware
// it takes exactly the v1.0.0 path: [demojifyPreserving] for
// AllowedEmojis, then [demojifyAllowed], then [Demojify].
func removeEmoji(text string, opts Options) string {
	switch {
	case opts.ScriptAware:
		return demojifyScriptAware(text, opts.AllowedRanges, allowedSpans(text, opts.AllowedEmojis))
	case len(opts.AllowedEmojis) > 0:
		return demojifyPreserving(text, opts.AllowedEmojis, opts.AllowedRanges)
	case len(opts.AllowedRanges) > 0:
		return demojifyAllowed(text, opts.AllowedRanges)
	default:
		return Demojify(text)
	}
}

// demojifyScriptAware removes emoji codepoints from text, keeping a ZWJ or
// a variation selector that is not part of an emoji sequence, any rune in
// allowed (as [demojifyAllowed] does), and every byte protected marks (the
// allowed emoji, from [allowedSpans]; nil protects nothing).
//
// It reads the text as written, protected bytes included, so a joiner or a
// selector beside an allowed emoji is classified by the emoji itself. (The
// placeholders [demojifyPreserving] swaps in would hide it: a joiner after
// an allowed emoji would look like a joiner after ordinary text, and stay.)
func demojifyScriptAware(text string, allowed []*unicode.RangeTable, protected []bool) string {
	locs := emojiRE.FindAllStringIndex(text, -1)
	if len(locs) == 0 {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	last := 0
	for _, l := range locs {
		if l[0] < len(protected) && protected[l[0]] {
			continue
		}
		r, _ := utf8.DecodeRuneInString(text[l[0]:l[1]])
		if len(allowed) > 0 && unicode.IsOneOf(allowed, r) {
			continue
		}
		if keepInScript(text, l[0], l[1], r) {
			continue
		}
		b.WriteString(text[last:l[0]])
		last = l[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

// keepInScript reports whether the codepoint r at text[start:end] is
// ordinary script text rather than part of an emoji sequence:
//
//   - a ZWJ with no emoji on either side (skipping variation selectors,
//     which sit between an emoji and its joiner);
//   - a standardized variation selector (U+FE00-U+FE0D) that does not
//     follow an emoji.
//
// The presentation selectors U+FE0E and U+FE0F exist only to choose text
// or emoji presentation, so they are removed wherever they appear, as by
// default; removing them leaves the base character (a copyright sign, a
// keycap digit) in its plain text form. Every other emoji codepoint is
// removed.
func keepInScript(text string, start, end int, r rune) bool {
	switch {
	case r == zwj:
		return !isEmojiBase(prevBase(text, start)) && !isEmojiBase(nextRune(text, end))
	case r >= variationFirst && r <= variationLastScript:
		return !isEmojiBase(prevRune(text, start))
	}
	return false
}

// isEmojiBase reports an emoji codepoint that is not itself a joiner, a
// variation selector or the combining keycap: a character an emoji
// sequence is built around.
func isEmojiBase(r rune) bool {
	if r == utf8.RuneError || r == zwj || r == combiningKeycap ||
		(r >= variationFirst && r <= presentationEmoji) {
		return false
	}
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return emojiRE.Match(buf[:n])
}

// prevRune is the rune ending at text[:i], or utf8.RuneError at the start.
func prevRune(text string, i int) rune {
	if i <= 0 {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeLastRuneInString(text[:i])
	return r
}

// prevBase is the rune before text[:i], skipping variation selectors, so a
// joiner after "U+2764 U+FE0F" (heart, emoji presentation) sees the heart.
func prevBase(text string, i int) rune {
	for i > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:i])
		if r >= variationFirst && r <= presentationEmoji {
			i -= size
			continue
		}
		return r
	}
	return utf8.RuneError
}

// nextRune is the rune starting at text[i:], or utf8.RuneError at the end.
func nextRune(text string, i int) rune {
	if i >= len(text) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeRuneInString(text[i:])
	return r
}

// hasNonEmpty reports whether ss holds a non-empty string; empty entries
// in Options.AllowedEmojis are ignored.
func hasNonEmpty(ss []string) bool {
	for _, s := range ss {
		if s != "" {
			return true
		}
	}
	return false
}

// findingHasEmoji is [Finding.HasEmoji]: [ContainsEmoji], or with
// opts.ScriptAware the script-aware answer, so a file whose only matches
// are joiners and variation selectors in ordinary text reports none.
func findingHasEmoji(text string, opts Options) bool {
	if opts.ScriptAware {
		return ContainsEmojiWith(text, Options{ScriptAware: true})
	}
	return ContainsEmoji(text)
}

// allowedSpans marks the bytes of text inside an occurrence of an allowed
// emoji, chosen the way [demojifyPreserving]'s placeholders choose them:
// longest strings first, and for each, every leftmost occurrence that does
// not overlap one already marked. It returns nil when there is nothing to
// protect (no non-empty entries, or none occurs). Empty entries are
// ignored, as Sanitize ignores them.
func allowedSpans(text string, allowedEmojis []string) []bool {
	sorted := make([]string, 0, len(allowedEmojis))
	for _, e := range allowedEmojis {
		if e != "" {
			sorted = append(sorted, e)
		}
	}
	if len(sorted) == 0 {
		return nil
	}
	sortByLenDesc(sorted)
	var mask []bool
	for _, e := range sorted {
		for i := 0; i <= len(text)-len(e); {
			j := strings.Index(text[i:], e)
			if j < 0 {
				break
			}
			start, end := i+j, i+j+len(e)
			if mask != nil && marked(mask[start:end]) {
				// It overlaps an occurrence already protected, so the
				// placeholder pass would not see it either.
				i = start + 1
				continue
			}
			if mask == nil {
				mask = make([]bool, len(text))
			}
			for k := start; k < end; k++ {
				mask[k] = true
			}
			i = end
		}
	}
	return mask
}

// marked reports whether any byte in m is marked.
func marked(m []bool) bool {
	for _, v := range m {
		if v {
			return true
		}
	}
	return false
}
