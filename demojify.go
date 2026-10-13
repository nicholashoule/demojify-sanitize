// Emoji detection and removal: the emojiRE character class, Demojify,
// ContainsEmoji, CountEmoji, BytesSaved, and TechnicalSymbolRanges.
//
// emojiRE is the single source of truth for what counts as emoji; every other
// file (sanitize.go, replace.go, scan.go) matches through it rather than
// defining ranges of its own. It matches one codepoint at a time, so ZWJ
// sequences and variation selectors are stripped piecewise. demojifyAllowed
// and demojifyPreserving implement Options.AllowedRanges and
// Options.AllowedEmojis; the latter protects the allowed occurrences by
// marking their bytes in place (allowedSpans), so no input text can collide
// with the mechanism.
// Range rationale: docs/unicode-coverage.md.

package demojify

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// emojiRE is a character class that matches individual emoji-related codepoints:
// pictographic symbols, Zero Width Joiner (U+200D), variation selectors
// (U+FE00–U+FE0F), and Combining Enclosing Keycap (U+20E3). Each codepoint is
// matched and removed independently -- ZWJ and variation selectors are stripped
// as isolated characters, not as part of a multi-codepoint sequence.
// Unicode ranges follow the Unicode 15–17 emoji/pictographic assignments.
//
// Ranges covered:
//
//	U+2139           Information Source
//	U+231A–U+231B   Watch, Hourglass Done
//	U+23CF           Eject Symbol
//	U+23E9–U+23F3   Fast-forward, hourglass, etc.
//	U+23F8–U+23FA   Pause, Stop, Record
//	U+24C2           Circled M
//	U+25AA–U+25AB   Small squares
//	U+25B6           Play button
//	U+25C0           Reverse button
//	U+25FB–U+25FE   Medium squares
//	U+2600–U+27BF   Miscellaneous Symbols + Dingbats
//	U+2934–U+2935   Curved arrows
//	U+2B05–U+2B07   Directional arrows
//	U+2B1B–U+2B1C   Large squares
//	U+2B50           Star
//	U+2B55           Circle
//	U+3030           Wavy dash
//	U+303D           Part alternation mark
//	U+3297           Circled Congratulation
//	U+3299           Circled Secret
//	U+1F000–U+1FAFF All supplementary emoji blocks (incl. Regional Indicators,
//	                 skin-tone modifiers, and new Emoji 17.0 additions)
//	U+200D           Zero Width Joiner (stripped as an individual codepoint)
//	U+20E3           Combining Enclosing Keycap
//	U+E0020–U+E007F  Tags block (subdivision flag tag sequences: England, Scotland, Wales)
//	U+FE00–U+FE0F   Variation Selectors 1–16
var emojiRE = regexp.MustCompile(
	`[\x{2139}` +
		`\x{231A}-\x{231B}` +
		`\x{23CF}` +
		`\x{23E9}-\x{23F3}` +
		`\x{23F8}-\x{23FA}` +
		`\x{24C2}` +
		`\x{25AA}-\x{25AB}` +
		`\x{25B6}` +
		`\x{25C0}` +
		`\x{25FB}-\x{25FE}` +
		`\x{2600}-\x{27BF}` +
		`\x{2934}-\x{2935}` +
		`\x{2B05}-\x{2B07}` +
		`\x{2B1B}-\x{2B1C}` +
		`\x{2B50}` +
		`\x{2B55}` +
		`\x{3030}` +
		`\x{303D}` +
		`\x{3297}` +
		`\x{3299}` +
		`\x{1F000}-\x{1FAFF}` +
		`\x{200D}` +
		`\x{20E3}` +
		`\x{E0020}-\x{E007F}` +
		`\x{FE00}-\x{FE0F}]`,
)

// Demojify removes emoji-related codepoints recognized by the package.
// Surrounding ASCII and non-emoji Unicode text is left unchanged.
func Demojify(text string) string {
	return emojiRE.ReplaceAllString(text, "")
}

// ContainsEmoji reports whether text contains at least one emoji or
// Unicode pictographic character recognized by [Demojify].
func ContainsEmoji(text string) bool {
	return emojiRE.MatchString(text)
}

// CountEmoji returns the number of emoji codepoint occurrences found in
// text. Each matched codepoint (including ZWJ, variation selectors, and
// other combining characters) counts as one occurrence.
// CountEmoji is safe for concurrent use.
func CountEmoji(text string) int {
	return len(emojiRE.FindAllString(text, -1))
}

// BytesSaved returns the number of bytes that [Demojify] would remove from
// text. It is equivalent to
// len(text) - len([Demojify](text)).
// BytesSaved is safe for concurrent use.
func BytesSaved(text string) int {
	return len(text) - len(Demojify(text))
}

// technicalSymbols is a Unicode range table covering technical symbols that
// overlap with the emojiRE character class but are not emoji clutter. These
// include check marks, ballot boxes, warning signs, gear symbols, card suits,
// stars, and music notation characters that LLMs commonly produce in
// structured output.
var technicalSymbols = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x2605, Hi: 0x2606, Stride: 1}, // Black/White Star
		{Lo: 0x2610, Hi: 0x2612, Stride: 1}, // Ballot Box variants
		{Lo: 0x2660, Hi: 0x2667, Stride: 1}, // Card suits
		{Lo: 0x266D, Hi: 0x266F, Stride: 1}, // Music notation
		{Lo: 0x2696, Hi: 0x2696, Stride: 1}, // Scales (balance)
		{Lo: 0x2699, Hi: 0x2699, Stride: 1}, // Gear
		{Lo: 0x269B, Hi: 0x269B, Stride: 1}, // Atom Symbol
		{Lo: 0x26A0, Hi: 0x26A0, Stride: 1}, // Warning Sign
		{Lo: 0x2713, Hi: 0x2714, Stride: 1}, // Check Marks
		{Lo: 0x2715, Hi: 0x2718, Stride: 1}, // Multiplication/Ballot X
		{Lo: 0x271A, Hi: 0x271A, Stride: 1}, // Heavy Greek Cross
	},
}

// TechnicalSymbolRanges returns Unicode range tables covering technical
// symbols commonly produced by LLMs that fall within the emoji regex range
// but are not emoji clutter. Pass the result to [Options.AllowedRanges] to
// preserve these symbols during sanitization:
//
//	opts := demojify.DefaultOptions()
//	opts.AllowedRanges = demojify.TechnicalSymbolRanges()
//	clean := demojify.Sanitize(text, opts)
//
// Covered symbols include check marks, ballot boxes, warning signs, gears,
// card suits, stars, and music notation characters.
func TechnicalSymbolRanges() []*unicode.RangeTable {
	return []*unicode.RangeTable{technicalSymbols}
}

// demojifyAllowed removes emoji codepoints from text while preserving any rune
// that belongs to at least one of the provided Unicode range tables. Callers
// pass this from [Sanitize] when len(opts.AllowedRanges) > 0 (a non-empty slice).
func demojifyAllowed(text string, allowed []*unicode.RangeTable) string {
	return emojiRE.ReplaceAllStringFunc(text, func(s string) string {
		r, _ := utf8.DecodeRuneInString(s)
		if unicode.IsOneOf(allowed, r) {
			return s
		}
		return ""
	})
}

// demojifyPreserving removes emoji codepoints from text while preserving
// specific emoji strings listed in allowedEmojis and any rune belonging to
// the provided Unicode range tables. Longer allowed strings are matched
// before shorter sub-sequences to handle variation-selector and ZWJ
// sequences correctly; each leftmost occurrence that does not overlap one
// already protected is kept (see [allowedSpans]).
//
// The occurrences are protected by marking their bytes in place. Earlier
// releases swapped Unicode noncharacter placeholders into the text and
// restored them afterwards; input that already held the placeholder
// characters (all 34 noncharacters plus the fallback sequence), or an
// allowed string that matched inside a placeholder (a digit), corrupted the
// output. Marking bytes has no such collision.
//
// Empty strings in allowedEmojis are ignored.
func demojifyPreserving(text string, allowedEmojis []string, allowedRanges []*unicode.RangeTable) string {
	return demojifyProtected(text, allowedRanges, allowedSpans(text, allowedEmojis))
}

// demojifyProtected removes every emoji codepoint from text except a rune in
// allowed and one starting at a byte protected marks (nil protects nothing).
// It is the v1 removal rule: a Zero Width Joiner or a variation selector is
// removed wherever it appears (see [demojifyScriptAware] for the
// script-aware rule).
func demojifyProtected(text string, allowed []*unicode.RangeTable, protected []bool) string {
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
		if len(allowed) > 0 {
			if r, _ := utf8.DecodeRuneInString(text[l[0]:l[1]]); unicode.IsOneOf(allowed, r) {
				continue
			}
		}
		b.WriteString(text[last:l[0]])
		last = l[1]
	}
	b.WriteString(text[last:])
	return b.String()
}
