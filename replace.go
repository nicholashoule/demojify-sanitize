package demojify

import (
	"os"
	"path/filepath"
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
// Replace is safe for concurrent use.
func Replace(text string, replacements map[string]string) string {
	if len(replacements) == 0 {
		return Demojify(text)
	}
	// Build a Replacer with keys ordered longest-first so that multi-codepoint
	// sequences (e.g., U+26A0 U+FE0F) are consumed before their component codepoints.
	keys := sortedKeys(replacements)
	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k, replacements[k])
	}
	text = strings.NewReplacer(args...).Replace(text)
	return Demojify(text)
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
// substitutions and removals performed (emoji codepoints from the replacement
// map plus any residual emoji codepoints stripped by [Demojify]).
//
// ReplaceFile returns an error for any filesystem failure. When count is zero
// the file is unchanged and no write is performed.
// ReplaceFile is safe for concurrent use provided callers do not share path.
func ReplaceFile(path string, replacements map[string]string) (count int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	original := string(data)
	cleaned := Replace(original, replacements)
	if cleaned == original {
		return 0, nil
	}

	count = countReplacements(original, replacements)

	// Preserve original permissions.
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	perm := info.Mode().Perm()

	// Atomic write: write to a sibling temp file, then rename.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".demojify-*")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()
	if _, err = tmp.WriteString(cleaned); err != nil {
		tmp.Close()
		return 0, err
	}
	if err = tmp.Close(); err != nil {
		return 0, err
	}
	if err = os.Chmod(tmpName, perm); err != nil {
		return 0, err
	}
	if err = os.Rename(tmpName, path); err != nil {
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
// FindAllMapped is safe for concurrent use.
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
// and the total number of substitutions and removals performed. It is
// equivalent to calling [Replace] and then counting replacements separately,
// but does both in a single pass.
//
// ReplaceCount is safe for concurrent use.
func ReplaceCount(text string, replacements map[string]string) (string, int) {
	cleaned := Replace(text, replacements)
	if cleaned == text {
		return text, 0
	}
	return cleaned, countReplacements(text, replacements)
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

// countReplacements returns the total number of positions in original that
// will be modified by Replace with the given replacements map. It counts each
// map key occurrence once (longest-first, mirroring Replace's behaviour) then
// adds the number of residual emoji codepoints that Demojify would remove.
func countReplacements(original string, replacements map[string]string) int {
	if len(replacements) == 0 {
		return len(emojiRE.FindAllString(original, -1))
	}
	keys := sortedKeys(replacements)
	count := 0
	remaining := original
	for _, k := range keys {
		n := strings.Count(remaining, k)
		count += n
		if n > 0 {
			remaining = strings.ReplaceAll(remaining, k, replacements[k])
		}
	}
	count += len(emojiRE.FindAllString(remaining, -1))
	return count
}
