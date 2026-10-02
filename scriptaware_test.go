package demojify

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every string here is built from code points so the source stays ASCII:
// this repository's own pre-commit hook strips emoji, joiners and
// variation selectors from files, which would silently rewrite the cases.

// cp builds a string from code points.
func cp(rs ...rune) string { return string(rs) }

const (
	tZWJ   = 0x200D
	tVS1   = 0xFE00
	tVS16  = 0xFE0F
	tKeycp = 0x20E3
)

var scriptAware = Options{RemoveEmojis: true, ScriptAware: true}

// Text that uses a joiner or a standardized variation selector as part of
// ordinary script, not emoji.
var scriptCases = []struct {
	name string
	text string
}{
	{"Devanagari half form (ka virama ZWJ ssa)", cp(0x0915, 0x094D, tZWJ, 0x0937)},
	{"Malayalam chillu (na virama ZWJ)", cp(0x0D28, 0x0D4D, tZWJ)},
	{"Sinhala yansaya (ka virama ZWJ ya)", cp(0x0D9A, 0x0DCA, tZWJ, 0x0DBA)},
	{"CJK standardized variant", cp(0x8279, tVS1)},
	{"math standardized variant", cp(0x2269, tVS1)},
	{"Indic word in a sentence", "Note: " + cp(0x0915, 0x094D, tZWJ, 0x0937) + " (ksha)."},
}

// ScriptAware keeps the joiner and the variation selector in ordinary
// script; the default still removes them (the v1.0.0 behavior, pinned).
func TestScriptAwareKeepsScriptText(t *testing.T) {
	t.Parallel()
	for _, tc := range scriptCases {
		if got := Sanitize(tc.text, scriptAware); got != tc.text {
			t.Errorf("%s: ScriptAware = %q, want it unchanged %q", tc.name, got, tc.text)
		}
		if ContainsEmojiWith(tc.text, scriptAware) {
			t.Errorf("%s: ContainsEmojiWith(ScriptAware) = true, want false", tc.name)
		}
		// The default is unchanged: v1.0.0 removes them, and existing
		// callers keep that.
		if Demojify(tc.text) == tc.text || !ContainsEmoji(tc.text) {
			t.Errorf("%s: the default behavior changed: Demojify = %q, ContainsEmoji = %v", tc.name, Demojify(tc.text), ContainsEmoji(tc.text))
		}
	}
}

// ScriptAware removes every emoji, every joiner between emoji, and the
// presentation selectors, exactly as the default does.
func TestScriptAwareRemovesEmojiSequences(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, in, want string
	}{
		{"single emoji", "a" + cp(0x1F600) + "b", "ab"},
		{"family ZWJ sequence", cp(0x1F468, tZWJ, 0x1F469, tZWJ, 0x1F467), ""},
		{"rainbow flag (VS16 then ZWJ)", cp(0x1F3F3, tVS16, tZWJ, 0x1F308), ""},
		{"heart on fire", cp(0x2764, tVS16, tZWJ, 0x1F525), ""},
		{"skin tone", cp(0x1F44D, 0x1F3FD), ""},
		{"flag", cp(0x1F1EF, 0x1F1F5), ""},
		{"keycap keeps its digit", cp('1', tVS16, tKeycp), "1"},
		{"emoji presentation of a text symbol", cp(0xA9, tVS16), cp(0xA9)},
		{"tag sequence", cp(0x1F3F4, 0xE0067, 0xE0062, 0xE0073, 0xE0063, 0xE0074, 0xE007F), ""},
		{"hidden tag characters", "hi" + cp(0xE0069, 0xE0067), "hi"},
		{"joiner before an emoji goes with it", cp(0x0915, 0x094D, tZWJ, 0x1F600), cp(0x0915, 0x094D)},
		{"text symbol", cp(0x2713) + " done", " done"},
	} {
		if got := Sanitize(tc.in, scriptAware); got != tc.want {
			t.Errorf("%s: ScriptAware = %q, want %q", tc.name, got, tc.want)
		}
		if !ContainsEmojiWith(tc.in, scriptAware) {
			t.Errorf("%s: ContainsEmojiWith(ScriptAware) = false, want true", tc.name)
		}
	}
}

// randomText draws from an alphabet built to stress the context rules:
// letters, Indic script, joiners, every kind of selector, emoji, keycaps
// and tag characters, in any order.
func randomText(r *rand.Rand) string {
	alphabet := []rune{
		'a', 'Z', '1', ' ', '\n', 0x0915, 0x094D, 0x0937, 0x0D28, 0x0D4D, 0x8279, 0x2269, 0xA9,
		tZWJ, tVS1, 0xFE0D, 0xFE0E, tVS16, tKeycp,
		0x1F600, 0x1F468, 0x1F3FD, 0x1F1EF, 0x2764, 0x2713, 0x2B05, 0x3030, 0xE0067, 0xE007F,
	}
	n := r.Intn(12)
	rs := make([]rune, n)
	for i := range rs {
		rs[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(rs)
}

// ScriptAware never removes anything the default keeps, and differs from
// it only by joiners and variation selectors it kept: removing what the
// default removes from its output gives the default's output.
func TestScriptAwareNeverRemovesMoreThanTheDefault(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		text := randomText(r)
		def, aware := Demojify(text), Sanitize(text, scriptAware)
		if Demojify(aware) != def {
			t.Fatalf("%q: ScriptAware kept %q, which the default does not reduce to %q", text, aware, def)
		}
		for _, k := range aware {
			if ContainsEmoji(string(k)) && k != tZWJ && (k < tVS1 || k > 0xFE0D) {
				t.Fatalf("%q: ScriptAware kept %U, which only the joiner and FE00-FE0D may be", text, k)
			}
		}
		// Detection agrees with removal.
		if ContainsEmojiWith(text, scriptAware) != (aware != text) {
			t.Fatalf("%q: ContainsEmojiWith = %v but Sanitize changed it: %v", text, ContainsEmojiWith(text, scriptAware), aware != text)
		}
		// A zero Options detects exactly as ContainsEmoji.
		if ContainsEmojiWith(text, Options{}) != ContainsEmoji(text) {
			t.Fatalf("%q: ContainsEmojiWith(zero Options) differs from ContainsEmoji", text)
		}
	}
}

// ContainsEmojiWith honors the allowed lists, as Sanitize does.
func TestContainsEmojiWithHonorsAllowedLists(t *testing.T) {
	t.Parallel()
	rocket, grin := cp(0x1F680), cp(0x1F600)
	if ContainsEmojiWith(cp(0x2713)+" passed", Options{AllowedRanges: TechnicalSymbolRanges()}) {
		t.Error("a check mark inside TechnicalSymbolRanges is not removed, so it is not reported")
	}
	if ContainsEmojiWith("go "+rocket, Options{AllowedEmojis: []string{rocket}}) {
		t.Error("an allowed emoji is not reported")
	}
	if !ContainsEmojiWith(rocket+grin, Options{AllowedEmojis: []string{rocket}}) {
		t.Error("an emoji outside the allowed list is reported")
	}
	if !ContainsEmojiWith(grin, Options{AllowedEmojis: []string{""}}) {
		t.Error("empty AllowedEmojis entries are ignored, as Sanitize ignores them")
	}
	withAllowed := Options{RemoveEmojis: true, ScriptAware: true, AllowedEmojis: []string{rocket}}
	text := rocket + " " + cp(0x0915, 0x094D, tZWJ, 0x0937) + grin
	if got, want := Sanitize(text, withAllowed), rocket+" "+cp(0x0915, 0x094D, tZWJ, 0x0937); got != want {
		t.Errorf("ScriptAware with AllowedEmojis = %q, want %q", got, want)
	}
}

// ScriptAware reaches every function built on Sanitize.
func TestScriptAwareThroughThePipeline(t *testing.T) {
	t.Parallel()
	indic := cp(0x0915, 0x094D, tZWJ, 0x0937)
	grin := cp(0x1F600)

	rep := SanitizeReport(indic+grin, scriptAware)
	if rep.Cleaned != indic || rep.EmojiRemoved != 1 {
		t.Errorf("SanitizeReport = %q removed %d, want %q removed 1", rep.Cleaned, rep.EmojiRemoved, indic)
	}

	js, err := SanitizeJSON([]byte(`{"name":"`+indic+grin+`","n":1}`), scriptAware)
	if err != nil || !bytes.Contains(js, []byte(indic)) || bytes.Contains(js, []byte(grin)) {
		t.Errorf("SanitizeJSON = %s, %v; want the Indic word kept and the emoji removed", js, err)
	}

	// SanitizeReader writes no newline after the last line (its v1.0.0
	// behavior).
	var out bytes.Buffer
	if err := SanitizeReader(strings.NewReader(indic+grin+"\n"+indic+"\n"), &out, scriptAware); err != nil ||
		out.String() != indic+"\n"+indic {
		t.Errorf("SanitizeReader = %q, %v", out.String(), err)
	}

	dir := t.TempDir()
	clean := filepath.Join(dir, "clean.md")
	mixed := filepath.Join(dir, "mixed.md")
	if err := os.WriteFile(clean, []byte("# Title\n\n"+indic+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mixed, []byte(indic+" "+grin+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if f, err := ScanFile(clean, scriptAware); err != nil || f != nil {
		t.Errorf("ScanFile(Indic only) = %+v, %v; want no finding", f, err)
	}
	f, err := ScanFile(mixed, scriptAware)
	if err != nil || f == nil || !f.HasEmoji || f.Cleaned != indic+" \n" {
		t.Errorf("ScanFile(mixed) = %+v, %v; want a finding that keeps the Indic word", f, err)
	}

	cfg := DefaultScanConfig()
	cfg.Root = dir
	cfg.Options = scriptAware
	findings, err := ScanDir(cfg)
	if err != nil || len(findings) != 1 || filepath.Base(findings[0].Path) != "mixed.md" {
		t.Errorf("ScanDir(ScriptAware) = %d findings, %v; want only mixed.md", len(findings), err)
	}
}

// A joiner or a selector beside an allowed emoji is classified by the
// emoji itself, not by the placeholder demojifyPreserving hides it behind
// (PR review of v1.1.0): the rule is the same with or without the
// allowed list.
func TestScriptAwareAllowedEmojiBoundaries(t *testing.T) {
	t.Parallel()
	rocket := cp(0x1F680)
	heart := cp(0x2764, tVS16)
	family := cp(0x1F468, tZWJ, 0x1F469)
	indic := cp(0x0915, 0x094D, tZWJ, 0x0937)
	for _, tc := range []struct {
		name    string
		in      string
		allowed []string
		want    string
	}{
		{"joiner after an allowed emoji", rocket + cp(tZWJ) + "abc", []string{rocket}, rocket + "abc"},
		{"joiner before an allowed emoji", "abc" + cp(tZWJ) + rocket, []string{rocket}, "abc" + rocket},
		{"selector after an allowed emoji", rocket + cp(tVS1), []string{rocket}, rocket},
		{"joiner after an allowed emoji in emoji form", heart + cp(tZWJ) + "x", []string{heart}, heart + "x"},
		{"Indic word beside an allowed emoji", indic + " " + rocket, []string{rocket}, indic + " " + rocket},
		{"allowed sequence kept whole", family + " " + cp(0x1F600), []string{family}, family + " "},
		{"empty entries ignored", indic + cp(0x1F600), []string{""}, indic},
	} {
		opts := Options{RemoveEmojis: true, ScriptAware: true, AllowedEmojis: tc.allowed}
		if got := Sanitize(tc.in, opts); got != tc.want {
			t.Errorf("%s: Sanitize = %q, want %q", tc.name, got, tc.want)
		}
		if got, want := ContainsEmojiWith(tc.in, opts), tc.want != tc.in; got != want {
			t.Errorf("%s: ContainsEmojiWith = %v, want %v", tc.name, got, want)
		}
	}
}

// With an allowed list, ScriptAware still differs from the default only by
// the joiners and selectors it keeps: the default (placeholder) removal
// applied to its output gives the default's output, and detection agrees
// with removal.
func TestScriptAwareWithAllowedEmojisMatchesTheDefault(t *testing.T) {
	t.Parallel()
	sets := [][]string{
		{cp(0x1F600)},
		{cp(0x2764, tVS16)},
		{cp(0x1F468, tZWJ, 0x1F600)},
		{"", cp(0x2713)},
		{cp(0x1F600), cp(0x1F468, 0x1F3FD)},
	}
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 20000; i++ {
		text := randomText(r)
		for _, allowed := range sets {
			plain := Options{RemoveEmojis: true, AllowedEmojis: allowed}
			aware := Options{RemoveEmojis: true, ScriptAware: true, AllowedEmojis: allowed}
			out := Sanitize(text, aware)
			if Sanitize(out, plain) != Sanitize(text, plain) {
				t.Fatalf("%q allowed %q: ScriptAware gave %q, which the default does not reduce to %q",
					text, allowed, out, Sanitize(text, plain))
			}
			if ContainsEmojiWith(text, aware) != (out != text) {
				t.Fatalf("%q allowed %q: ContainsEmojiWith disagrees with Sanitize", text, allowed)
			}
		}
	}
}
