// Native fuzz tests for the sanitizer paths changed in v1.2.0: the
// recommended SanitizeSafe path and the AllowedEmojis protection, checked
// against the pre-v1.2.0 placeholder implementation where that one is
// correct. Run beyond the seed corpus with, e.g.:
//
//	go test -run '^$' -fuzz FuzzSanitizeSafe -fuzztime 30s .

package demojify

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// fuzzSeeds covers emoji families, flags, keycaps, presentation and
// standardized selectors, Indic joiners, noncharacters and invalid UTF-8.
var fuzzSeeds = []string{
	"",
	"plain ASCII text",
	"ok \U0001F600 done",
	"\u0915\u094D\u200D\u0937 \U0001F600",
	"\U0001F468\u200D\U0001F469\u200D\U0001F467",
	"\u2764\uFE0F\u200D\U0001F525",
	"\U0001F1FA\U0001F1F8 \U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F",
	"1\uFE0F\u20E3 #\u20E3",
	"\u8279\uFE00 \u2205\uFE00",
	"\u200D\u200D\uFE0F\uFE0E",
	"\uFDD0\uFDD1\uFDD20\uFDD0\uFDD1\uFDD2\U0001F680",
	"\xF0\x9F\U0001F600\x98\x80",
	"a\xffb\U0001F680c",
}

// FuzzSanitizeSafe: the recommended path only ever deletes, leaves nothing
// SafeOptions would still remove, is idempotent, and keeps valid UTF-8 valid.
func FuzzSanitizeSafe(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out := SanitizeSafe(in)
		if len(out) > len(in) {
			t.Fatalf("output grew: %q -> %q", in, out)
		}
		if utf8.ValidString(in) {
			if !utf8.ValidString(out) {
				t.Fatalf("valid input produced invalid UTF-8: %q -> %q", in, out)
			}
			if ContainsEmojiWith(out, SafeOptions()) {
				t.Fatalf("emoji left behind: %q -> %q", in, out)
			}
			if again := SanitizeSafe(out); again != out {
				t.Fatalf("not idempotent: %q -> %q -> %q", in, out, again)
			}
		}
	})
}

// legacyDemojifyPreserving is the pre-v1.2.0 AllowedEmojis implementation
// (noncharacter placeholders swapped in and restored), kept as the reference
// FuzzAllowedEmojis compares against on input it handled correctly.
func legacyDemojifyPreserving(text string, allowedEmojis []string) string {
	sorted := make([]string, 0, len(allowedEmojis))
	for _, e := range allowedEmojis {
		if e != "" {
			sorted = append(sorted, e)
		}
	}
	if len(sorted) == 0 {
		return Demojify(text)
	}
	sortByLenDesc(sorted)
	prefix := string(rune(0xFDD0))
	placeholders := make([]string, len(sorted))
	for i := range sorted {
		placeholders[i] = prefix + strconv.Itoa(i) + prefix
	}
	protected := text
	for i, e := range sorted {
		protected = strings.ReplaceAll(protected, e, placeholders[i])
	}
	cleaned := Demojify(protected)
	for i, e := range sorted {
		cleaned = strings.ReplaceAll(cleaned, placeholders[i], e)
	}
	return cleaned
}

// hasPlaceholderHazard reports text the legacy implementation mishandled: a
// noncharacter (its sentinels) or an ASCII digit (its placeholder indices).
func hasPlaceholderHazard(s string) bool {
	for _, r := range s {
		if (r >= 0xFDD0 && r <= 0xFDEF) || r == 0xFFFE || r == 0xFFFF || (r >= '0' && r <= '9') {
			return true
		}
	}
	return false
}

// FuzzAllowedEmojis: AllowedEmojis only ever keeps emoji the caller allowed
// (removing every emoji from the result and from the input gives the same
// text), and matches the legacy placeholder implementation wherever that one
// had no collision. allowed is split on "|" into the allow list; equal-length
// entries are sorted unstably, so the differential uses distinct lengths.
func FuzzAllowedEmojis(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s, "\U0001F680|\u2764\uFE0F")
		f.Add(s+" \U0001F680\U0001F680", "\U0001F680\U0001F680|\U0001F680|1")
	}
	f.Fuzz(func(t *testing.T, text, allowed string) {
		if !utf8.ValidString(text) || !utf8.ValidString(allowed) {
			return
		}
		list := strings.Split(allowed, "|")
		got := demojifyPreserving(text, list, nil)
		if Demojify(got) != Demojify(text) {
			t.Fatalf("removed more than emoji: %q allowed %q -> %q", text, list, got)
		}
		if aware := demojifyScriptAware(text, nil, allowedSpans(text, list)); len(aware) < len(got) {
			t.Fatalf("script-aware removed more than v1: %q allowed %q: %q vs %q", text, list, aware, got)
		}
		lengths := map[int]bool{}
		for _, e := range list {
			if e != "" && lengths[len(e)] {
				return
			}
			lengths[len(e)] = true
		}
		if hasPlaceholderHazard(text) || hasPlaceholderHazard(allowed) {
			return
		}
		if want := legacyDemojifyPreserving(text, list); got != want {
			t.Fatalf("differs from the legacy implementation: %q allowed %q: got %q, want %q", text, list, got, want)
		}
	})
}
