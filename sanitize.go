package demojify

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
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
	// RemoveEmojis is false or the slice is nil or empty.
	AllowedRanges []*unicode.RangeTable

	// AllowedEmojis lists specific emoji strings that are preserved during
	// emoji removal. Unlike [AllowedRanges], which preserves entire Unicode
	// blocks, AllowedEmojis targets individual emoji -- including
	// multi-codepoint sequences such as ZWJ family emoji or flag sequences.
	// Longer strings are matched before shorter sub-sequences. Has no effect
	// when RemoveEmojis is false or the slice is nil or empty.
	AllowedEmojis []string
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
		switch {
		case len(opts.AllowedEmojis) > 0:
			text = demojifyPreserving(text, opts.AllowedEmojis, opts.AllowedRanges)
		case len(opts.AllowedRanges) > 0:
			text = demojifyAllowed(text, opts.AllowedRanges)
		default:
			text = Demojify(text)
		}
	}
	if opts.NormalizeWhitespace {
		text = Normalize(text)
	}
	return text
}

// SanitizeFile reads the file at path, applies [Sanitize] with opts, and
// writes the result back only if changes were made. The original file
// permissions are preserved and a temp-file-plus-rename strategy is used
// for safe writes (see [ReplaceFile] for platform details).
//
// Binary files (detected by a NUL byte in the first 512 bytes) are silently
// skipped and return (false, nil), matching the behavior of [ScanDir] and
// [ScanFile].
//
// SanitizeFile returns true when the file was modified and false when it
// was already clean. It returns an error for any filesystem failure.
func SanitizeFile(path string, opts Options) (changed bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if isBinary(data) {
		return false, nil
	}
	original := string(data)
	cleaned := Sanitize(original, opts)
	if cleaned == original {
		return false, nil
	}
	return true, statAndWrite(path, cleaned)
}

// SanitizeResult holds the cleaned output and metrics from a sanitization
// pass. It is returned by [SanitizeReport] for agent pipelines that need
// observability: audit trails, cost-impact logging, or conditional routing
// based on change volume.
type SanitizeResult struct {
	// Cleaned is the sanitized text.
	Cleaned string

	// EmojiRemoved is the number of emoji codepoint occurrences that were
	// removed or replaced during sanitization. Zero when
	// [Options.RemoveEmojis] is false.
	EmojiRemoved int

	// BytesSaved is the number of bytes saved by the full sanitization
	// pipeline (emoji removal plus whitespace normalization if enabled).
	BytesSaved int
}

// SanitizeReport applies the same pipeline as [Sanitize] and returns
// structured metrics alongside the cleaned text. The
// [SanitizeResult.EmojiRemoved] field is zero when opts.RemoveEmojis
// is false.
func SanitizeReport(text string, opts Options) SanitizeResult {
	cleaned := Sanitize(text, opts)
	emojiCount := 0
	if opts.RemoveEmojis {
		emojiCount = CountEmoji(text)
	}
	return SanitizeResult{
		Cleaned:      cleaned,
		EmojiRemoved: emojiCount,
		BytesSaved:   len(text) - len(cleaned),
	}
}

// SanitizeReader reads text from r, applies the sanitization pipeline
// defined by opts, and writes the cleaned result to w line by line. It is
// designed for streaming scenarios (e.g., processing LLM token streams or
// MCP transport payloads) where buffering the complete input is
// undesirable.
//
// When opts.NormalizeWhitespace is true, each line is normalized
// individually: trailing whitespace is trimmed, inline space runs are
// collapsed, and runs of consecutive blank lines are limited to one.
// Leading and trailing blank lines are omitted, closely matching
// [Normalize]. Unlike [Normalize], bare CR (\r) mid-line is not converted
// because [bufio.Scanner] only splits on \n (bare CR is rare in practice).
//
// SanitizeReader returns an error for any I/O failure or scanner error.
func SanitizeReader(r io.Reader, w io.Writer, opts Options) error {
	scanner := bufio.NewScanner(r)
	wroteAny := false
	pendingBlanks := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Step 1: emoji removal.
		if opts.RemoveEmojis {
			switch {
			case len(opts.AllowedEmojis) > 0:
				line = demojifyPreserving(line, opts.AllowedEmojis, opts.AllowedRanges)
			case len(opts.AllowedRanges) > 0:
				line = demojifyAllowed(line, opts.AllowedRanges)
			default:
				line = Demojify(line)
			}
		}

		// Step 2: per-line whitespace normalization.
		if opts.NormalizeWhitespace {
			line = collapseInlineSpaces(line)
			line = strings.TrimRight(line, " \t")

			if line == "" {
				if wroteAny {
					pendingBlanks++
				}
				continue
			}

			// Flush at most one pending blank line before this
			// non-blank line.
			if pendingBlanks > 0 {
				if _, err := io.WriteString(w, "\n\n"); err != nil {
					return err
				}
				pendingBlanks = 0
			} else if wroteAny {
				if _, err := io.WriteString(w, "\n"); err != nil {
					return err
				}
			}
		} else if wroteAny {
			// No normalization: separate lines with newlines.
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
		wroteAny = true
	}

	return scanner.Err()
}

// SanitizeJSON sanitizes only the string values within a JSON document,
// leaving keys, numbers, booleans, and null values unchanged. This
// prevents the corruption of JSON structure that would occur if [Sanitize]
// were applied to raw JSON bytes.
//
// SanitizeJSON preserves numeric precision by decoding numbers as
// json.Number tokens. The returned bytes are compact JSON (no
// indentation).
//
// SanitizeJSON returns an error if data is not valid JSON.
func SanitizeJSON(data []byte, opts Options) ([]byte, error) {
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	v = sanitizeJSONValue(v, opts)
	return json.Marshal(v)
}

// sanitizeJSONValue recursively sanitizes string values in a decoded JSON
// structure, returning the sanitized value. Non-string leaf values
// (numbers, booleans, null) are returned unchanged.
func sanitizeJSONValue(v interface{}, opts Options) interface{} {
	switch val := v.(type) {
	case string:
		return Sanitize(val, opts)
	case map[string]interface{}:
		for k, child := range val {
			val[k] = sanitizeJSONValue(child, opts)
		}
		return val
	case []interface{}:
		for i, child := range val {
			val[i] = sanitizeJSONValue(child, opts)
		}
		return val
	default:
		return v
	}
}
