// The Sanitize pipeline and its entry points: Options, DefaultOptions,
// Sanitize, SanitizeFile, SanitizeReport, SanitizeReader, and SanitizeJSON.
//
// Order is fixed: emoji removal (honoring AllowedEmojis, then AllowedRanges)
// followed by whitespace normalization. SanitizeReader applies the same steps
// per line with a 1 MiB line cap; SanitizeJSON applies them to string values
// only and rejects multi-value input with ErrMultipleJSONValues.
// Replacement maps are not consulted here; see replace.go and scan.go.

package demojify

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"unicode"
)

// ErrMultipleJSONValues is returned by [SanitizeJSON] when the input contains
// more than one top-level JSON value (e.g., `{"a":1}{"b":2}` or `{"a":1} 5`).
var ErrMultipleJSONValues = errors.New("demojify: input contains multiple JSON values")

// Options configures the sanitization pipeline used by [Sanitize].
//
// The zero value disables every step: [Sanitize] with Options{} returns its
// input unchanged. That is the stable v1 behavior, kept for existing
// callers; it is easy to get wrong, because an Options a caller forgot to
// fill in sanitizes nothing. New code should start from [SafeOptions] (or
// call [SanitizeSafe]), which always removes emoji and keeps script joiners.
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

	// ScriptAware keeps the Zero Width Joiner (U+200D) and the standardized
	// variation selectors (U+FE00-U+FE0D) where they are not part of an
	// emoji sequence. Devanagari, Malayalam, Sinhala and other Indic
	// scripts join letters with ZWJ, and CJK, Mongolian and mathematical
	// text use standardized variation sequences; without ScriptAware those
	// codepoints are removed wherever they appear, which changes how such
	// text renders. Every emoji, every joiner between emoji, and the
	// presentation selectors U+FE0E and U+FE0F are removed as before.
	//
	// Applies to [Sanitize] and the functions built on it
	// ([SanitizeReport], [SanitizeReader], [SanitizeJSON], [SanitizeFile],
	// the scanner's Options) and to [ContainsEmojiWith]. [Demojify],
	// [ContainsEmoji], [CountEmoji] and the [Replace] family are unchanged.
	// For sanitization, has no effect when RemoveEmojis is false;
	// [ContainsEmojiWith] honors it regardless. Added in v1.1.0; off by
	// default (the stable v1 default), on in [SafeOptions].
	ScriptAware bool
}

// DefaultOptions returns an Options value with emoji removal and whitespace
// normalization enabled. ScriptAware stays off, the v1.0.0 behavior, so a
// Zero Width Joiner or a variation selector is removed wherever it appears,
// including inside Indic, CJK and mathematical text; prefer [SafeOptions]
// (adding NormalizeWhitespace when wanted) in new code.
func DefaultOptions() Options {
	return Options{
		RemoveEmojis:        true,
		NormalizeWhitespace: true,
	}
}

// SafeOptions returns the recommended Options: emoji removal on, with
// ScriptAware, and whitespace left alone. Every emoji, every joiner between
// emoji and the presentation selectors are removed; a Zero Width Joiner or a
// standardized variation selector that joins or varies script text stays.
// Leaving whitespace alone makes it safe for any text, including the string
// values [SanitizeJSON] visits and source code; set NormalizeWhitespace on
// the returned value to collapse whitespace too.
//
// Use it with [Sanitize], [SanitizeJSON], [SanitizeReader], [SanitizeFile]
// and [ContainsEmojiWith], or call [SanitizeSafe]. Added in v1.2.0.
func SafeOptions() Options {
	return Options{
		RemoveEmojis: true,
		ScriptAware:  true,
	}
}

// SanitizeSafe is [Sanitize] with [SafeOptions]: it always removes emoji,
// keeps script joiners, and leaves whitespace alone. It is the recommended
// entry point for new code, since no Options value can be left unset by
// mistake. Added in v1.2.0.
func SanitizeSafe(text string) string {
	return Sanitize(text, SafeOptions())
}

// Sanitize applies the sanitization steps defined in opts to text and returns
// the cleaned result. Steps are applied in the following order:
//
//  1. Emoji removal ([Demojify]) when opts.RemoveEmojis is true, honoring
//     opts.AllowedEmojis, opts.AllowedRanges and opts.ScriptAware.
//  2. Whitespace normalization ([Normalize]) when opts.NormalizeWhitespace is true.
func Sanitize(text string, opts Options) string {
	if opts.RemoveEmojis {
		text = removeEmoji(text, opts)
	}
	if opts.NormalizeWhitespace {
		text = Normalize(text)
	}
	return text
}

// SanitizeFile reads the file at path, applies [Sanitize] with opts, and
// writes the result back only if changes were made. The original file
// permissions are preserved and a same-directory temp-file-plus-rename
// strategy is used (atomic on POSIX, best-effort replacement on Windows).
//
// Binary files (detected by a NUL byte in the first 512 bytes) are silently
// skipped and return (false, nil), matching the behavior of [ScanDir] and
// [ScanFile].
//
// The write happens only if the file still holds what was read: a file
// another writer changed in between is left alone and SanitizeFile returns
// (false, err) with err wrapping [ErrFileChanged].
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
	if err := statAndWrite(path, original, cleaned); err != nil {
		return false, err
	}
	return true, nil
}

// SanitizeResult holds the cleaned output and metrics from a sanitization
// pass. It is returned by [SanitizeReport] for agent pipelines that need
// observability: audit trails, cost-impact logging, or conditional routing
// based on change volume.
type SanitizeResult struct {
	// Cleaned is the sanitized text.
	Cleaned string

	// EmojiRemoved is the number of recognized emoji-related codepoint
	// occurrences removed during sanitization. Zero when
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
		emojiCount = CountEmoji(text) - CountEmoji(cleaned)
	}
	return SanitizeResult{
		Cleaned:      cleaned,
		EmojiRemoved: emojiCount,
		BytesSaved:   len(text) - len(cleaned),
	}
}

// sanitizeReaderMaxTokenSize is the maximum line length (in bytes) that
// [SanitizeReader] will process. Lines longer than this limit cause
// [bufio.ErrTooLong] to be returned. 1 MiB accommodates minified JSON,
// base64-encoded payloads, and long LLM output lines.
const sanitizeReaderMaxTokenSize = 1024 * 1024 // 1 MiB per line

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
// Lines up to 1 MiB (sanitizeReaderMaxTokenSize) are supported. Longer
// lines cause [bufio.ErrTooLong] to be returned.
//
// The output never ends with a newline: lines are written separated by
// newlines, so a trailing newline on the final input line is not
// reproduced. This holds even when opts.NormalizeWhitespace is false and
// matches the trailing-trim behavior of [Normalize].
//
// SanitizeReader returns an error for any I/O failure or scanner error.
func SanitizeReader(r io.Reader, w io.Writer, opts Options) error {
	scanner := bufio.NewScanner(r)
	// Scanner needs room beyond the content limit to observe an LF or CRLF
	// delimiter. The explicit length check below enforces the content limit.
	scanner.Buffer(make([]byte, 0, 64*1024), sanitizeReaderMaxTokenSize+2)
	wroteAny := false
	pendingBlanks := 0

	for scanner.Scan() {
		if len(scanner.Bytes()) > sanitizeReaderMaxTokenSize {
			return bufio.ErrTooLong
		}
		line := scanner.Text()

		// Step 1: emoji removal.
		if opts.RemoveEmojis {
			line = removeEmoji(line, opts)
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
// SanitizeJSON returns an error if data is not valid JSON or if it
// contains trailing non-whitespace content after the first value (e.g.,
// `{"a":1} trailing`).
func SanitizeJSON(data []byte, opts Options) ([]byte, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	// Ensure the entire input is a single valid JSON value by requiring EOF
	// after the first successful decode. Decoding into json.RawMessage
	// accepts ANY well-formed second value (object, array, number, string,
	// bool, null), so every multiple-value input consistently returns
	// [ErrMultipleJSONValues]; trailing garbage that is not valid JSON
	// (e.g. `{"a":1} trailing`) returns the decoder's syntax error instead.
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, ErrMultipleJSONValues
		}
		return nil, err
	}
	v = sanitizeJSONValue(v, opts)
	return json.Marshal(v)
}

// sanitizeJSONValue recursively sanitizes string values in a decoded JSON
// structure, returning the sanitized value. Non-string leaf values
// (numbers, booleans, null) are returned unchanged.
func sanitizeJSONValue(v any, opts Options) any {
	switch val := v.(type) {
	case string:
		return Sanitize(val, opts)
	case map[string]any:
		for k, child := range val {
			val[k] = sanitizeJSONValue(child, opts)
		}
		return val
	case []any:
		for i, child := range val {
			val[i] = sanitizeJSONValue(child, opts)
		}
		return val
	default:
		return v
	}
}
