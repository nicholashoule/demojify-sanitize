// Command driver demonstrates the primary features of the demojify-sanitize
// module. It exercises emoji detection, removal, substitution, whitespace
// normalization, and directory scanning with realistic input.
//
// Run:
//
//	go run ./docs/examples/driver
package main

import (
	"fmt"
	"os"
	"path/filepath"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func main() {
	// ---- 1. Detect emoji presence ----
	fmt.Println("=== ContainsEmoji ===")
	samples := []string{
		"\u2705 build passed",
		"plain text only",
		"\U0001F680 deploy complete",
	}
	for _, s := range samples {
		fmt.Printf("  ContainsEmoji(%q) = %v\n", s, demojify.ContainsEmoji(s))
	}

	// ---- 2. Strip all emoji ----
	fmt.Println("\n=== Demojify ===")
	raw := "\U0001F680 Deploy complete! Check \U0001F4CA for details."
	fmt.Printf("  before: %q\n", raw)
	fmt.Printf("  after:  %q\n", demojify.Demojify(raw))

	// ---- 3. Normalize whitespace ----
	fmt.Println("\n=== Normalize ===")
	messy := "Hello   World  \n\n\n\nMore text"
	fmt.Printf("  before: %q\n", messy)
	fmt.Printf("  after:  %q\n", demojify.Normalize(messy))

	// ---- 4. Full pipeline via Sanitize ----
	fmt.Println("\n=== Sanitize (DefaultOptions) ===")
	input := "\U0001F680 Deploy complete!\n\n\nCheck the logs \U0001F4CA"
	clean := demojify.Sanitize(input, demojify.DefaultOptions())
	fmt.Printf("  before: %q\n", input)
	fmt.Printf("  after:  %q\n", clean)

	// ---- 5. Substitution with DefaultReplacements ----
	fmt.Println("\n=== Replace (DefaultReplacements) ===")
	repl := demojify.DefaultReplacements()
	text := "\u2705 tests passed, \u274c build failed, \u26a0 review needed"
	fmt.Printf("  before: %q\n", text)
	result := demojify.Replace(text, repl)
	fmt.Printf("  after:  %q\n", result)

	// ---- 6. ReplaceCount -- substitute and count ----
	fmt.Println("\n=== ReplaceCount ===")
	rc, n := demojify.ReplaceCount("\u2705 OK \u274c FAIL \U0001F680 deploy", repl)
	fmt.Printf("  result: %q (%d substitution(s))\n", rc, n)

	// ---- 7. FindAll -- discover distinct emoji ----
	fmt.Println("\n=== FindAll ===")
	found := demojify.FindAll("\u2705 pass \u274c fail \u2705 again")
	fmt.Printf("  found: %v\n", found) // prints literal emoji sequences

	// ---- 8. FindAllMapped -- only mapped keys ----
	fmt.Println("\n=== FindAllMapped ===")
	mapped := demojify.FindAllMapped("\u2705 pass \U0001F600 smile", repl)
	fmt.Printf("  mapped keys: %v\n", mapped) // only emoji with entries in repl

	// ---- 9. ScanDir -- audit a directory ----
	fmt.Println("\n=== ScanDir ===")
	tmpDir, err := os.MkdirTemp("", "demojify-driver-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "MkdirTemp: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Populate temp files.
	writeFile(tmpDir, "status.md", "\u2705 All tests passed\n")
	writeFile(tmpDir, "clean.txt", "No emoji here\n")
	writeFile(tmpDir, "deploy.log", "\U0001F680 deployed to prod\n")

	cfg := demojify.DefaultScanConfig()
	cfg.Root = tmpDir
	cfg.CollectMatches = true

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ScanDir: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  %d file(s) with findings\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  - %s (has_emoji=%v, matches=%d)\n", f.Path, f.HasEmoji, len(f.Matches))
		for _, m := range f.Matches {
			display := m.Replacement
			if display == "" {
				display = "(stripped)"
			}
			fmt.Printf("      line %d col %d: %q -> %s\n", m.Line, m.Column, m.Sequence, display)
		}
	}

	// ---- 10. WriteFinding -- atomic write-back ----
	fmt.Println("\n=== WriteFinding ===")
	for _, f := range findings {
		absPath := filepath.Join(tmpDir, filepath.FromSlash(f.Path))
		changed, werr := demojify.WriteFinding(absPath, f)
		if werr != nil {
			fmt.Fprintf(os.Stderr, "  WriteFinding(%s): %v\n", f.Path, werr)
			continue
		}
		fmt.Printf("  %s: changed=%v\n", f.Path, changed)
	}

	// Verify files are clean after write-back.
	findings2, _ := demojify.ScanDir(cfg)
	fmt.Printf("  after fix: %d remaining finding(s)\n", len(findings2))

	// ---- 11. FixDir -- scan and write back in one call ----
	fmt.Println("\n=== FixDir ===")
	fixDir, err := os.MkdirTemp("", "demojify-fixdir-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "MkdirTemp: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(fixDir)

	writeFile(fixDir, "dirty1.md", "\u2705 All tests passed\n")
	writeFile(fixDir, "dirty2.md", "\U0001F680 deployed to prod\n")
	writeFile(fixDir, "clean.md", "No emoji here\n")

	fixCfg := demojify.DefaultScanConfig()
	fixed, _, fixErr := demojify.FixDir(fixDir, fixCfg)
	if fixErr != nil {
		fmt.Fprintf(os.Stderr, "FixDir: %v\n", fixErr)
		os.Exit(1)
	}
	fmt.Printf("  fixed %d file(s)\n", fixed)

	// Verify idempotency -- second run should find nothing.
	fixed2, _, _ := demojify.FixDir(fixDir, fixCfg)
	fmt.Printf("  idempotent re-run: %d file(s) fixed\n", fixed2)

	fmt.Println("\n[PASS] driver completed successfully")
}

func writeFile(dir, name, content string) {
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "WriteFile %s: %v\n", name, err)
		os.Exit(1)
	}
}
