package demojify

import "regexp"

// emojiRE matches individual emoji code points including ZWJ sequences,
// variation selectors, and combining enclosing keycaps. Unicode ranges
// follow the Unicode 15 emoji/pictographic assignments.
//
// Ranges covered:
//
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
//	U+1F000–U+1FAFF All supplementary emoji blocks
//	U+200D           Zero Width Joiner (joins sequences)
//	U+20E3           Combining Enclosing Keycap
//	U+FE00–U+FE0F   Variation Selectors 1–16
var emojiRE = regexp.MustCompile(
	`[\x{231A}-\x{231B}` +
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
		`\x{FE00}-\x{FE0F}]`,
)

// Demojify removes emoji and Unicode pictographic characters from text,
// replacing each matched code point with an empty string. Surrounding
// ASCII and non-emoji Unicode text is left unchanged.
func Demojify(text string) string {
	return emojiRE.ReplaceAllString(text, "")
}

// ContainsEmoji reports whether text contains at least one emoji or
// Unicode pictographic character recognised by [Demojify].
func ContainsEmoji(text string) bool {
	return emojiRE.MatchString(text)
}
