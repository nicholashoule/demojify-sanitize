package demojify

import (
	"os"
	"path/filepath"
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

// SanitizeFile reads the file at path, applies [Sanitize] with opts, and
// writes the result back only if changes were made. The original file
// permissions are preserved and a temp-file-plus-rename strategy is used
// for safe writes (see [ReplaceFile] for platform details).
//
// SanitizeFile returns true when the file was modified and false when it
// was already clean. It returns an error for any filesystem failure.
func SanitizeFile(path string, opts Options) (changed bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	original := string(data)
	cleaned := Sanitize(original, opts)
	if cleaned == original {
		return false, nil
	}

	// Preserve original permissions.
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	perm := info.Mode().Perm()

	// Atomic write: write to a sibling temp file, then rename.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".demojify-*")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()
	if _, err = tmp.WriteString(cleaned); err != nil {
		tmp.Close()
		return false, err
	}
	if err = tmp.Close(); err != nil {
		return false, err
	}
	if err = os.Chmod(tmpName, perm); err != nil {
		return false, err
	}
	if err = os.Rename(tmpName, path); err != nil {
		return false, err
	}
	return true, nil
}
