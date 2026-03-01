package demojify

import (
	"os"
	"sort"
	"strings"
)

// sortedKeys returns the keys of m sorted by descending byte length so that
// longer sequences are matched before shorter sub-sequences (e.g., a key
// containing a variation selector such as U+FE0F is tried before its base
// codepoint).
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})
	return keys
}

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
// via [Demojify]. Called by [Replace], [ReplaceCount], and [ReplaceFile]
// to avoid re-sorting keys in composed operations.
func applyReplacer(text string, replacements map[string]string, keys []string) string {
	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k, replacements[k])
	}
	return Demojify(strings.NewReplacer(args...).Replace(text))
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
// ReplaceFile returns an error for any filesystem failure. When count is zero
// the file is unchanged and no write is performed.
// ReplaceFile is safe for concurrent use provided callers do not share path
// and the replacements map is not mutated concurrently.
func ReplaceFile(path string, replacements map[string]string) (count int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
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

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if err = atomicWrite(path, cleaned, info.Mode().Perm()); err != nil {
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
// matching the greedy behaviour of [Replace] (e.g., U+26A0 U+FE0F wins over
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

// FindMatchesInFile reads the file at path and returns a [Match] for every
// emoji codepoint found, with [Match.Replacement] populated from the
// replacements map (empty string if the codepoint is not mapped). Matches
// are ordered by line then column. Returns nil and no error when the file
// contains no emoji.
//
// Unlike [ScanDir] with CollectMatches, this function does not filter or
// sanitize the file; it only collects match metadata.
//
// FindMatchesInFile returns an error for any filesystem failure.
func FindMatchesInFile(path string, replacements map[string]string) ([]Match, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return buildMatches(string(data), replacements), nil
}

// countWithKeys performs a single left-to-right scan over text and returns the
// number of emoji positions: mapped-key matches (longest first, greedy) plus
// unmapped emoji codepoints found by [emojiRE]. This mirrors the matching
// behaviour of [Replace] without building intermediate strings for each key.
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
		if loc := emojiRE.FindStringIndex(text[i:]); loc != nil && loc[0] == 0 {
			count++
			i += loc[1]
		} else {
			i++
		}
	}
	return count
}
