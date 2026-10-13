// Tests for the v1.2.0 additions: SafeOptions, SanitizeSafe, SafeScanConfig
// and DefaultMaxFileBytes, plus the stable v1 defaults they sit beside.

package demojify_test

import (
	"os"
	"path/filepath"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// hindiKsha is Devanagari "ksha" with a Zero Width Joiner for its half
// form: script text, not an emoji sequence.
const hindiKsha = "\u0915\u094D\u200D\u0937"

func TestSafeOptions(t *testing.T) {
	got := demojify.SafeOptions()
	if !got.RemoveEmojis || !got.ScriptAware || got.NormalizeWhitespace ||
		got.AllowedRanges != nil || got.AllowedEmojis != nil {
		t.Errorf("SafeOptions() = %+v, want RemoveEmojis and ScriptAware only", got)
	}
}

func TestSanitizeSafe(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"empty", "", ""},
		{"plain ASCII", "hello world", "hello world"},
		{"single emoji", "ok \U0001F600", "ok "},
		{"script joiner kept", hindiKsha + " \U0001F600", hindiKsha + " "},
		{"ZWJ family removed whole", "a\U0001F468\u200D\U0001F469\u200D\U0001F467b", "ab"},
		{"presentation selector removed", "\u2764\uFE0F!", "!"},
		{"standardized variation kept", "\u8279\uFE00", "\u8279\uFE00"},
		{"flag removed", "\U0001F1FA\U0001F1F8 US", " US"},
		{"keycap removed to digit", "1\uFE0F\u20E3", "1"},
		{"whitespace untouched", "a  \U0001F600  b\n\n\n", "a    b\n\n\n"},
		{"JSON-shaped text untouched but emoji", "{\"k\": \"v \u2705\"}", `{"k": "v "}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := demojify.SanitizeSafe(tt.in)
			if got != tt.want {
				t.Errorf("SanitizeSafe(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if again := demojify.Sanitize(tt.in, demojify.SafeOptions()); again != got {
				t.Errorf("Sanitize(SafeOptions) = %q, SanitizeSafe = %q; want equal", again, got)
			}
			if demojify.ContainsEmojiWith(got, demojify.SafeOptions()) {
				t.Errorf("SanitizeSafe(%q) = %q still holds emoji under SafeOptions", tt.in, got)
			}
		})
	}
}

// TestStableV1Defaults pins the v1 behavior SafeOptions exists beside, so a
// change to it is a deliberate, major-version decision: a zero Options passes
// text through, and DefaultOptions removes a script joiner.
func TestStableV1Defaults(t *testing.T) {
	tests := []struct {
		name string
		opts demojify.Options
		in   string
		want string
	}{
		{"zero Options passes through", demojify.Options{}, "x \U0001F600", "x \U0001F600"},
		{"DefaultOptions is not script-aware", demojify.DefaultOptions(), hindiKsha, "\u0915\u094D\u0937"},
		{"SafeOptions is", demojify.SafeOptions(), hindiKsha, hindiKsha},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := demojify.Sanitize(tt.in, tt.opts); got != tt.want {
				t.Errorf("Sanitize(%q, %+v) = %q, want %q", tt.in, tt.opts, got, tt.want)
			}
		})
	}
	if !demojify.ContainsEmoji(hindiKsha) {
		t.Error("ContainsEmoji(hindi) = false; the v1 detector flags a lone joiner")
	}
	if demojify.ContainsEmojiWith(hindiKsha, demojify.SafeOptions()) {
		t.Error("ContainsEmojiWith(hindi, SafeOptions) = true, want false")
	}
}

func TestSafeScanConfig(t *testing.T) {
	cfg := demojify.SafeScanConfig()
	def := demojify.DefaultScanConfig()
	if cfg.Options.RemoveEmojis != true || cfg.Options.ScriptAware != true {
		t.Errorf("SafeScanConfig().Options = %+v, want SafeOptions()", cfg.Options)
	}
	if cfg.MaxFileBytes != demojify.DefaultMaxFileBytes || def.MaxFileBytes != demojify.DefaultMaxFileBytes ||
		demojify.DefaultMaxFileBytes != 1<<20 {
		t.Errorf("MaxFileBytes safe=%d default=%d const=%d, want 1 MiB everywhere",
			cfg.MaxFileBytes, def.MaxFileBytes, demojify.DefaultMaxFileBytes)
	}
	if len(cfg.SkipExtensions) != len(def.SkipExtensions) || cfg.Root != def.Root {
		t.Errorf("SafeScanConfig filters differ from DefaultScanConfig")
	}

	// Only the file with a real emoji is a finding; script text is not.
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("hindi.md", hindiKsha+"\n")
	write("emoji.md", "done \u2705\n")
	cfg.Root = dir
	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Path != "emoji.md" {
		t.Errorf("ScanDir(SafeScanConfig) = %+v, want only emoji.md", findings)
	}
}
