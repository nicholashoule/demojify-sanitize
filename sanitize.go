package demojify

import (
	"unicode"
)

// Options configures the sanitization pipeline used by [Sanitize].
// Zero value disables all steps; use [DefaultOptions] for sensible defaults.
type Options struct {
	// RemoveEmojis strips emoji and Unicode pictographic characters.
	// Equivalent to calling [Demojify] on the input.
	RemoveEmojis bool

	// NormalizeWhitespace collapses redundant spaces and blank lines.
	// Equivalent to calling [Normalize] on the input.
	NormalizeWhitespace bool

	// AllowedRanges lists Unicode ranges whose codepoints are preserved during
	// emoji removal. A codepoint that would normally be stripped by [Demojify]
	// is kept when it belongs to any table in this slice. Has no effect when
	// RemoveEmojis is false or the slice is nil.
	AllowedRanges []*unicode.RangeTable
}

// DefaultOptions returns an Options value with all sanitization steps enabled.
func DefaultOptions() Options {
	return Options{
		RemoveEmojis:        true,
		NormalizeWhitespace: true,
	}
}

// Sanitize applies the sanitization steps defined in opts to text and returns
// the cleaned result. Steps are applied in the following order:
//
//  1. Emoji removal ([Demojify]) when opts.RemoveEmojis is true.
//  2. Whitespace normalization ([Normalize]) when opts.NormalizeWhitespace is true.
func Sanitize(text string, opts Options) string {
	if opts.RemoveEmojis {
		if len(opts.AllowedRanges) == 0 {
			text = Demojify(text)
		} else {
			text = demojifyAllowed(text, opts.AllowedRanges)
		}
	}
	if opts.NormalizeWhitespace {
		text = Normalize(text)
	}
	return text
}
