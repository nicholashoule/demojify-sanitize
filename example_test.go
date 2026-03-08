package demojify_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func ExampleDemojify() {
	fmt.Println(demojify.Demojify("\U0001F680 Deploy complete! Check the logs \U0001F4CA"))
	// Output:
	//  Deploy complete! Check the logs
}

func ExampleContainsEmoji() {
	fmt.Println(demojify.ContainsEmoji("Hello \U0001F600 World"))
	fmt.Println(demojify.ContainsEmoji("Hello World"))
	// Output:
	// true
	// false
}

func ExampleNormalize() {
	fmt.Println(demojify.Normalize("Hello   World\n\n\nMore text"))
	// Output:
	// Hello World
	//
	// More text
}

func ExampleSanitize() {
	input := "\U0001F680 Deploy complete!\n\n\nCheck the logs \U0001F4CA"
	fmt.Println(demojify.Sanitize(input, demojify.DefaultOptions()))
	// Output:
	// Deploy complete!
	//
	// Check the logs
}

func ExampleSanitize_selective() {
	// Only remove emojis, leave whitespace untouched.
	opts := demojify.Options{RemoveEmojis: true}
	fmt.Println(demojify.Sanitize("Sure! \U0001F389 Done.", opts))
	// Output:
	// Sure!  Done.
}

// ExampleContainsEmoji_contentGate shows how to use ContainsEmoji as a
// guard before persisting or forwarding user-submitted text.
func ExampleContainsEmoji_contentGate() {
	report := "Q3 results: Revenue up 12% \U0001F4C8"

	if demojify.ContainsEmoji(report) {
		// Strip emojis and normalize before storing.
		opts := demojify.Options{RemoveEmojis: true, NormalizeWhitespace: true}
		fmt.Println(demojify.Sanitize(report, opts))
	}
	// Output:
	// Q3 results: Revenue up 12%
}

// ExampleSanitize_httpHandler shows how to wrap Sanitize in an HTTP handler
// that cleans an incoming plain-text request body before processing it.
// This example is compiled but not executed (no Output comment).
func ExampleSanitize_httpHandler() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		clean := demojify.Sanitize(string(body), demojify.DefaultOptions())
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, clean)
	})
	_ = handler
}

// ExampleSanitize_markdownFiles shows how to sanitize a known set of files
// using the Sanitize API directly -- useful when callers manage the file list
// themselves (e.g., only staged files from git diff --name-only).
// For whole-directory sanitization in one call, prefer ExampleFixDir.
// This example is compiled but not executed (no Output comment).
func ExampleSanitize_markdownFiles() {
	paths := []string{"README.md", "CHANGELOG.md", "CONTRIBUTING.md"}
	opts := demojify.DefaultOptions()

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			log.Printf("read %s: %v", p, err)
			continue
		}
		clean := demojify.Sanitize(string(data), opts)
		if err := os.WriteFile(p, []byte(clean), 0o644); err != nil {
			log.Printf("write %s: %v", p, err)
		}
	}
}

func ExampleDefaultOptions() {
	opts := demojify.DefaultOptions()
	fmt.Println(opts.RemoveEmojis, opts.NormalizeWhitespace)
	// Output:
	// true true
}

func ExampleDefaultReplacements() {
	repl := demojify.DefaultReplacements()
	fmt.Println(repl["\u2705"])
	fmt.Println(repl["\u274c"])
	fmt.Println(repl["\U0001F680"])
	// Output:
	// [PASS]
	// [FAIL]
	// [DEPLOY]
}

func ExampleDefaultScanConfig() {
	cfg := demojify.DefaultScanConfig()
	fmt.Println("Root:", cfg.Root)
	fmt.Println("SkipDirs:", cfg.SkipDirs)
	fmt.Println("ExemptSuffixes:", cfg.ExemptSuffixes)
	fmt.Println("MaxFileBytes:", cfg.MaxFileBytes)
	// Output:
	// Root: .
	// SkipDirs: [.git/ vendor/ node_modules/]
	// ExemptSuffixes: [_test.go]
	// MaxFileBytes: 1048576
}

func ExampleFindAll() {
	text := "build \u2705 done, launch \U0001F680 complete, check again \u2705"
	fmt.Println(demojify.FindAll(text))
	// Output:
	// [✅ 🚀]
}

func ExampleReplace() {
	repl := map[string]string{
		"\u2705": "[PASS]",
		"\u274c": "[FAIL]",
		"\u26a0": "WARNING",
	}
	fmt.Println(demojify.Replace("\u2705 tests passed, \u274c build failed", repl))
	// Output:
	// [PASS] tests passed, [FAIL] build failed
}

// ExampleReplace_defaultReplacements shows how to use the built-in
// substitution map so callers do not need to maintain their own.
func ExampleReplace_defaultReplacements() {
	out := demojify.Replace("\u2705 tests passed, \u274c build failed, \u26a0 review needed", demojify.DefaultReplacements())
	fmt.Println(out)
	// Output:
	// [PASS] tests passed, [FAIL] build failed, WARNING review needed
}

// ExampleReplaceFile shows how to apply a substitution map to a file in place.
// This example is compiled but not executed (no Output comment).
func ExampleReplaceFile() {
	repl := map[string]string{
		"\u2705":     "[PASS]",
		"\u274c":     "[FAIL]",
		"\u26a0":     "WARNING",
		"\U0001F680": "[DEPLOY]",
	}
	n, err := demojify.ReplaceFile("output.log", repl)
	if err != nil {
		log.Printf("ReplaceFile: %v", err)
		return
	}
	if n > 0 {
		log.Printf("replaced %d emoji occurrence(s) in output.log", n)
	}
}

func ExampleReplaceCount() {
	repl := demojify.DefaultReplacements()
	clean, n := demojify.ReplaceCount("\u2705 build OK, \u274c deploy failed", repl)
	fmt.Printf("%s (%d change(s))\n", clean, n)
	// Output:
	// [PASS] build OK, [FAIL] deploy failed (2 change(s))
}

func ExampleFindAllMapped() {
	repl := demojify.DefaultReplacements()
	fmt.Println(demojify.FindAllMapped("\u2705 pass \u274c fail \u26a0 warn", repl))
	// Output:
	// [✅ ❌ ⚠]
}

// ExampleFindMatchesInFile shows how to get per-line match detail from a file.
// This example is compiled but not executed (no Output comment).
func ExampleFindMatchesInFile() {
	repl := demojify.DefaultReplacements()
	matches, err := demojify.FindMatchesInFile("output.log", repl)
	if err != nil {
		log.Printf("FindMatchesInFile: %v", err)
		return
	}
	for _, m := range matches {
		if m.Replacement != "" {
			log.Printf("line %d col %d: %q -> %q", m.Line, m.Column, m.Sequence, m.Replacement)
		} else {
			log.Printf("line %d col %d: %q (unmapped -- will be stripped)", m.Line, m.Column, m.Sequence)
		}
	}
}

// ExampleScanDir shows how to audit a directory tree for emoji and inspect
// per-file findings.
func ExampleScanDir() {
	dir, err := os.MkdirTemp("", "example-scandir-*")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "status.md"), []byte("\u2705 done\n"), 0o644); err != nil {
		fmt.Println("error:", err)
		return
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = dir
	cfg.Extensions = []string{".md"}
	cfg.CollectMatches = true
	cfg.Replacements = demojify.DefaultReplacements()

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, f := range findings {
		fmt.Printf("found %d match(es), has_emoji=%v\n", len(f.Matches), f.HasEmoji)
	}
	// Output:
	// found 1 match(es), has_emoji=true
}

func ExampleScanFile() {
	f, err := demojify.ScanFile("README.md", demojify.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if f == nil {
		fmt.Println("clean")
	} else {
		fmt.Println("needs sanitization")
	}
	// (output depends on README.md content)
}

// ExampleSanitizeFile shows how to sanitize a single file in place.
// This example is compiled but not executed (no Output comment).
func ExampleSanitizeFile() {
	changed, err := demojify.SanitizeFile("output.md", demojify.DefaultOptions())
	if err != nil {
		log.Printf("SanitizeFile: %v", err)
		return
	}
	if changed {
		log.Println("output.md was sanitized")
	}
}

func ExampleWriteFinding() {
	dir, err := os.MkdirTemp("", "example-writefinding-*")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "status.md")
	if err := os.WriteFile(path, []byte("\u2705 done\n"), 0o644); err != nil {
		fmt.Println("error:", err)
		return
	}

	f, err := demojify.ScanFile(path, demojify.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if f == nil {
		fmt.Println("already clean")
		return
	}

	changed, err := demojify.WriteFinding(path, *f)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("changed:", changed)

	data, _ := os.ReadFile(path)
	fmt.Println("content:", string(data))
	// Output:
	// changed: true
	// content: done
}

// ExampleFixDir shows how to scan and fix an entire directory tree in one
// call. It writes back every file whose content changed and reports the
// number of files fixed and the number that were already clean.
func ExampleFixDir() {
	dir, err := os.MkdirTemp("", "example-fixdir-*")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer os.RemoveAll(dir)

	// Seed two files: one dirty, one clean.
	if err := os.WriteFile(filepath.Join(dir, "dirty.md"), []byte("launch \U0001F680\n"), 0o644); err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "clean.md"), []byte("all good\n"), 0o644); err != nil {
		fmt.Println("error:", err)
		return
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Extensions = []string{".md"}

	fixed, clean, err := demojify.FixDir(dir, cfg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("fixed=%d clean=%d\n", fixed, clean)

	// Verify the dirty file is now emoji-free.
	data, _ := os.ReadFile(filepath.Join(dir, "dirty.md"))
	fmt.Printf("dirty.md: %s", data)
	// Output:
	// fixed=1 clean=1
	// dirty.md: launch
}

func ExampleCountEmoji() {
	fmt.Println(demojify.CountEmoji("build \u2705 done, launch \U0001F680 complete"))
	// Output:
	// 2
}

func ExampleBytesSaved() {
	fmt.Println(demojify.BytesSaved("Hello \U0001F680 World"))
	// Output:
	// 4
}

func ExampleTechnicalSymbolRanges() {
	opts := demojify.Options{
		RemoveEmojis:  true,
		AllowedRanges: demojify.TechnicalSymbolRanges(),
	}
	fmt.Println(demojify.Sanitize("warning \u26A0 rocket \U0001F680", opts))
	// Output:
	// warning ⚠ rocket
}

func ExampleSanitizeReport() {
	result := demojify.SanitizeReport("Hello \U0001F600 World \U0001F680", demojify.DefaultOptions())
	fmt.Printf("cleaned=%q emoji=%d saved=%d\n", result.Cleaned, result.EmojiRemoved, result.BytesSaved)
	// Output:
	// cleaned="Hello World" emoji=2 saved=10
}

func ExampleSanitizeReader() {
	input := strings.NewReader("Hello \U0001F680 World\nLine 2 \u2705")
	var buf bytes.Buffer
	if err := demojify.SanitizeReader(input, &buf, demojify.Options{RemoveEmojis: true}); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(buf.String())
	// Output:
	// Hello  World
	// Line 2
}

func ExampleSanitizeJSON() {
	data := []byte(`{"status":"done \u2705","count":42}`)
	clean, err := demojify.SanitizeJSON(data, demojify.Options{RemoveEmojis: true})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(clean))
	// Output:
	// {"count":42,"status":"done "}
}

func ExampleScanDirContext() {
	dir, err := os.MkdirTemp("", "example-scandirctx-*")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "status.md"), []byte("\u2705 done\n"), 0o644); err != nil {
		fmt.Println("error:", err)
		return
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = dir
	cfg.Extensions = []string{".md"}

	findings, err := demojify.ScanDirContext(context.Background(), cfg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("found %d file(s) with emoji\n", len(findings))
	// Output:
	// found 1 file(s) with emoji
}
