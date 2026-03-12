package demojify

import (
	"os"
	"strings"
)

// Replace substitutes emoji codepoints found in text using the provided
// replacements map, then strips any remaining unmatched emoji codepoints
// via [Demojify]. Longer map keys are matched before shorter ones to handle
// variation-selector sequences (e.g., WARNING sign U+26A0 with U+FE0F) correctly.
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

// applyReplacer substitutes emoji codepoints using keys (pre-sorted by
// descending length) and the replacements map, then strips residual emoji
// via [Demojify]. Consecutive identical tokens produced by adjacent repeated
// emoji (e.g. two warning signs yielding "WARNINGWARNING") are collapsed to a
// single token. Called by [Replace], [ReplaceCount], and [ReplaceFile]
// to avoid re-sorting keys in composed operations.
func applyReplacer(text string, replacements map[string]string, keys []string) string {
	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k, replacements[k])
	}
	result := Demojify(strings.NewReplacer(args...).Replace(text))
	return collapseRepeatedTokens(result, distinctValues(replacements))
}

// distinctValues returns a deduplicated slice of the non-empty values in m,
// sorted by descending length so that longer tokens are collapsed before any
// shorter sub-string tokens they might contain.
func distinctValues(m map[string]string) []string {
	seen := make(map[string]struct{}, len(m))
	vals := make([]string, 0, len(m))
	for _, v := range m {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			vals = append(vals, v)
		}
	}
	sortByLenDesc(vals)
	return vals
}

// collapseRepeatedTokens reduces runs of the same replacement token in text
// to a single occurrence. Both direct concatenation ("TOKENTOK") and
// space-separated repetition ("TOKEN TOKEN") are collapsed, handling runs of
// three or more via repeated passes. Tokens in vals must be non-empty.
func collapseRepeatedTokens(text string, vals []string) string {
	for _, v := range vals {
		// Collapse space-separated repeats first so the concat pass can later
		// catch any newly adjacent duplicates.
		doubled := v + " " + v
		for strings.Contains(text, doubled) {
			text = strings.ReplaceAll(text, doubled, v)
		}
		// Collapse direct concatenation ("TOKTOKEN" -> "TOK").
		concat := v + v
		for strings.Contains(text, concat) {
			text = strings.ReplaceAll(text, concat, v)
		}
	}
	return text
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
// map plus any residual emoji codepoints stripped by [Demojify]).
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
	// Process text left-to-right with the same longest-first greedy match that
	// strings.NewReplacer uses, so variation-selector sequences are attributed
	// to the longer key rather than the bare codepoint.
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
// matches plus residual emoji stripped by [Demojify]).
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
