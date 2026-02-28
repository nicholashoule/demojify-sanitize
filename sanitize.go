package demojify

import "regexp"

// Options configures the sanitization pipeline used by [Sanitize].
// Zero value disables all steps; use [DefaultOptions] for sensible defaults.
type Options struct {
	// RemoveEmojis strips emoji and Unicode pictographic characters.
	// Equivalent to calling [Demojify] on the input.
	RemoveEmojis bool

	// RemoveAIClutter strips common AI-generated filler phrases such as
	// "Certainly!", "Sure,", and "I'd be happy to help!" when they appear
	// at the start of a line followed by punctuation.
	RemoveAIClutter bool

	// NormalizeWhitespace collapses redundant spaces and blank lines.
	// Equivalent to calling [Normalize] on the input.
	NormalizeWhitespace bool
}

// DefaultOptions returns an Options value with all sanitization steps enabled.
func DefaultOptions() Options {
	return Options{
		RemoveEmojis:        true,
		RemoveAIClutter:     true,
		NormalizeWhitespace: true,
	}
}

// aiClutterRE matches common AI-generated filler phrases at the start of a
// line. Unambiguous short words (Certainly, Sure, etc.) require trailing
// punctuation ([!,.]) to reduce false positives on legitimate text such as
// "Sure enough, …". Longer, structurally distinctive phrases allow optional
// punctuation. The optional trailing [ \t]*\n? consumes the whitespace and
// newline so callers do not need to tidy up afterwards.
var aiClutterRE = regexp.MustCompile(
	`(?im)^(` +
		`Certainly[!,.]|` +
		`Sure[!,.]|` +
		`Of\s+course[!,.]|` +
		`Absolutely[!,.]|` +
		`Great[!,.]|` +
		`Excellent[!,.]|` +
		`Noted[!,.]|` +
		`I'd be happy to(?: help)?[.!]?|` +
		`I can(?: certainly)? help(?: with that)?[.!]?|` +
		`I'll help you with that[.!]?|` +
		`Let me help you[.!]?|` +
		`I hope this helps[.!]?|` +
		`Feel free to ask if you need (?:more |further )?(?:help|assistance)[.!]?` +
		`)[ \t]*\n?`,
)

// Sanitize applies the sanitization steps defined in opts to text and returns
// the cleaned result. Steps are applied in the following order:
//
//  1. Emoji removal ([Demojify]) when opts.RemoveEmojis is true.
//  2. AI-clutter removal when opts.RemoveAIClutter is true.
//  3. Whitespace normalization ([Normalize]) when opts.NormalizeWhitespace is true.
func Sanitize(text string, opts Options) string {
	if opts.RemoveEmojis {
		text = Demojify(text)
	}
	if opts.RemoveAIClutter {
		text = aiClutterRE.ReplaceAllString(text, "")
	}
	if opts.NormalizeWhitespace {
		text = Normalize(text)
	}
	return text
}
