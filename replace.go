package demojify

import (
	"os"
	"strings"
	"unicode/utf8"
)

// minCollapseLen is the minimum replacement-token length eligible for
// run collapsing in [applyReplacer]. Shorter tokens (e.g. "-", "/", "*",
// "->", "<=") are common ASCII sequences that appear legitimately in source
// code, URLs, and documentation, and must never be deduplicated. Only
// label-like tokens such as "[FAIL]" or "WARNING" are safe to collapse.
const minCollapseLen = 4

// Replace substitutes emoji codepoints found in text using the provided
// replacements map, then strips any remaining unmatched emoji codepoints.
// Longer map keys are matched before shorter ones to handle
// variation-selector sequences (e.g., WARNING sign U+26A0 with U+FE0F) correctly.
//
// Replacement values are emitted verbatim: a value is never re-scanned for
// emoji, so identity mappings (key == value) preserve the mapped codepoints
// exactly. Runs of adjacent emoji that substitute to the same token of
// [minCollapseLen] or more bytes are collapsed to a single occurrence;
// literal text that happens to equal a replacement token is never collapsed.
//
// Replace with a nil or empty replacements map behaves identically to [Demojify].
// Replace is safe for concurrent use provided the replacements map is not
// mutated concurrently.
func Replace(text string, replacements map[string]string) string {
	if len(replacements) == 0 {
		return Demojify(text)
	}
	return applyReplacer(text, replacements, sortedKeys(replacements))
}

// applyReplacer walks text left to right, substituting mapped keys (tried
// longest-first at each position) and stripping unmapped emoji codepoints.
// Because it tracks which output spans were produced by substitution, it can
// collapse runs of adjacent substituted tokens without ever touching literal
// text that happens to equal a replacement token -- the position-blind
// find-and-replace it superseded corrupted such text. Called by [Replace],
// [ReplaceCount], [ReplaceFile], and the [ScanDir] walk to avoid re-sorting
// keys in composed operations.
//
// The keys parameter must hold the map's non-empty keys sorted by descending
// byte length (see [sortedKeys]).
func applyReplacer(text string, replacements map[string]string, keys []string) string {
	// Group keys by first byte, preserving longest-first order inside each
	// bucket, so the per-position candidate loop only sees keys that can
	// possibly match. ASCII text then skips the key loop entirely for
	// buckets with no keys.
	var byFirst [256][]string
	for _, k := range keys {
		byFirst[k[0]] = append(byFirst[k[0]], k)
	}

	out := make([]byte, 0, len(text))
	for i := 0; i < len(text); {
		c := text[i]
		matched := false
		for _, k := range byFirst[c] {
			if strings.HasPrefix(text[i:], k) {
				out = emitToken(out, replacements[k])
				i += len(k)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// Unmapped emoji are stripped. Every codepoint matched by emojiRE is
		// U+200D or higher, so its UTF-8 encoding starts with a non-ASCII
		// byte; ASCII bytes skip the regex entirely. Non-emoji runes are
		// copied whole so the regex runs once per rune, never on mid-rune
		// slices. Invalid UTF-8 decodes with size 1 and is copied through
		// byte-for-byte unchanged.
		if c >= utf8.RuneSelf {
			if loc := emojiRE.FindStringIndex(text[i:]); len(loc) > 0 && loc[0] == 0 {
				i += loc[1]
				continue
			}
			_, size := utf8.DecodeRuneInString(text[i:])
			out = append(out, text[i:i+size]...)
			i += size
			continue
		}
		out = append(out, c)
		i++
	}
	return string(out)
}

// emitToken appends a substituted token to out, collapsing runs: when out
// already ends with the same token -- directly or separated by a single
// space -- the new occurrence is dropped (and the separating space removed)
// so that adjacent repeated emoji yield one token instead of
// "TOKENTOKEN" or "TOKEN TOKEN". Only tokens of [minCollapseLen] or more
// bytes are collapsed; short tokens like "-" or "->" always append.
//
// Collapsing only ever suppresses the token being emitted, so literal input
// text is never removed: a run collapses only when its last member came from
// a substitution.
func emitToken(out []byte, tok string) []byte {
	if tok == "" {
		return out
	}
	if len(tok) >= minCollapseLen {
		if endsWith(out, tok) {
			return out
		}
		if n := len(out) - 1; n >= 0 && out[n] == ' ' && endsWith(out[:n], tok) {
			return out[:n]
		}
	}
	return append(out, tok...)
}

// endsWith reports whether b ends with the bytes of s. The string
// conversion inside the comparison does not allocate: the compiler
// recognizes the string(b)==s pattern and emits a direct memory compare
// (escape analysis: "string(...) does not escape"; verified zero allocs
// per call via testing.AllocsPerRun).
func endsWith(b []byte, s string) bool {
	return len(b) >= len(s) && string(b[len(b)-len(s):]) == s
}

// FindAll returns the distinct emoji codepoint sequences found in text.
// Each sequence appears at most once regardless of how many times it occurs.
// Sequences are returned in order of first occurrence.
// FindAll is safe for concurrent use.
func FindAll(text string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, seq := range emojiRE.FindAllString(text, -1) {
		if _, ok := seen[seq]; !ok {
			seen[seq] = struct{}{}
			result = append(result, seq)
		}
	}
	return result
}

// ReplaceFile reads the file at path, applies [Replace] with the provided
// replacements map, and writes the result back only if changes were made.
// The original file permissions are preserved. Returns the number of
// substitutions and removals performed (mapped sequences from the replacement
// map plus any residual unmapped emoji codepoints stripped).
//
// Binary files (detected by a NUL byte in the first 512 bytes) are silently
// skipped and return (0, nil), matching the behavior of [ScanDir] and
// [ScanFile].
//
// ReplaceFile returns an error for any filesystem failure. When count is zero
// the file is unchanged and no write is performed.
// ReplaceFile is safe for concurrent use provided callers do not share path
// and the replacements map is not mutated concurrently.
func ReplaceFile(path string, replacements map[string]string) (count int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if isBinary(data) {
		return 0, nil
	}
	original := string(data)

	var cleaned string
	var keys []string
	if len(replacements) > 0 {
		keys = sortedKeys(replacements)
		cleaned = applyReplacer(original, replacements, keys)
	} else {
		cleaned = Demojify(original)
	}
	if cleaned == original {
		return 0, nil
	}

	if len(keys) > 0 {
		count = countWithKeys(original, keys)
	} else {
		count = len(emojiRE.FindAllString(original, -1))
	}

	if err := statAndWrite(path, cleaned); err != nil {
		return 0, err
	}
	return count, nil
}

// FindAllMapped returns the distinct keys from replacements that appear in text,
// ordered by their first byte position in text. Only keys present in both text
// and the replacements map are returned; emoji codepoints not in the map are
// ignored. Use [FindAll] to find all emoji regardless of any map.
//
// Longer keys take priority over shorter sub-sequences at the same position,
// matching the greedy behavior of [Replace] (e.g., U+26A0 U+FE0F wins over
// bare U+26A0 when both are in the map and the text contains the full sequence).
//
// FindAllMapped is safe for concurrent use provided the replacements map is
// not mutated concurrently.
func FindAllMapped(text string, replacements map[string]string) []string {
	if len(replacements) == 0 || text == "" {
		return nil
	}
	// Process text left-to-right with the same longest-first greedy walk as
	// applyReplacer, so variation-selector sequences are attributed to the
	// longer key rather than the bare codepoint.
	keys := sortedKeys(replacements) // longest first
	seen := make(map[string]struct{})
	var result []string
	for i := 0; i < len(text); {
		matched := false
		for _, k := range keys {
			if strings.HasPrefix(text[i:], k) {
				if _, ok := seen[k]; !ok {
					seen[k] = struct{}{}
					result = append(result, k)
				}
				i += len(k)
				matched = true
				break
			}
		}
		if !matched {
			i++
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// ReplaceCount applies [Replace] to text and returns both the cleaned string
// and the total number of substitutions and removals performed (mapped-key
// matches plus residual unmapped emoji stripped). When adjacent repeated
// emoji collapse to a single token, each occurrence still counts.
//
// ReplaceCount is safe for concurrent use provided the replacements map is
// not mutated concurrently.
func ReplaceCount(text string, replacements map[string]string) (string, int) {
	if len(replacements) == 0 {
		cleaned := Demojify(text)
		if cleaned == text {
			return text, 0
		}
		return cleaned, len(emojiRE.FindAllString(text, -1))
	}
	keys := sortedKeys(replacements)
	cleaned := applyReplacer(text, replacements, keys)
	if cleaned == text {
		return text, 0
	}
	return cleaned, countWithKeys(text, keys)
}

// countWithKeys performs a single left-to-right scan over text and returns the
// number of emoji positions: mapped-key matches (longest first, greedy) plus
// unmapped emoji codepoints found by [emojiRE]. This mirrors the matching
// behavior of [Replace] without building intermediate strings for each key.
func countWithKeys(text string, keys []string) int {
	count := 0
	for i := 0; i < len(text); {
		matched := false
		for _, k := range keys {
			if strings.HasPrefix(text[i:], k) {
				count++
				i += len(k)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if loc := emojiRE.FindStringIndex(text[i:]); len(loc) > 0 && loc[0] == 0 {
			count++
			i += loc[1]
		} else {
			i++
		}
	}
	return count
}
