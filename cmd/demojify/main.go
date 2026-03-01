// Command demojify audits a directory tree for emoji, reports every occurrence
// with file, line, and column, and optionally rewrites affected files.
//
// Usage:
//
//	go run github.com/nicholashoule/demojify-sanitize/cmd/demojify [flags]
//
// Flags:
//
//	-root <dir>      directory to scan (default: ".")
//	-fix             rewrite affected files in place after reporting
//	-sub             substitute emoji with text tokens instead of stripping;
//	                 implies -fix
//	-exts <.go,.md>  comma-separated extensions to scan (default: all files)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func main() {
	root := flag.String("root", ".", "directory to scan")
	fix := flag.Bool("fix", false, "rewrite affected files in place")
	sub := flag.Bool("sub", false, "substitute emoji with text tokens (implies -fix)")
	exts := flag.String("exts", "", "comma-separated extensions to scan, e.g. .go,.md (default: all)")
	flag.Parse()

	if *sub {
		*fix = true
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = *root
	cfg.CollectMatches = true

	if *exts != "" {
		for _, e := range strings.Split(*exts, ",") {
			if e = strings.TrimSpace(e); e != "" {
				cfg.Extensions = append(cfg.Extensions, e)
			}
		}
	}

	repl := demojify.DefaultReplacements()
	if *sub {
		cfg.Replacements = repl
	}

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		log.Fatalf("scan error: %v", err)
	}

	if len(findings) == 0 {
		fmt.Println("[PASS] no emoji found")
		return
	}

	exitCode := 0
	for _, f := range findings {
		fmt.Printf("\n[WARN] %s\n", f.Path)
		for _, m := range f.Matches {
			display := m.Replacement
			if display == "" {
				display = "(stripped)"
			}
			fmt.Printf("  line %d col %d: %q -> %q\n", m.Line, m.Column, m.Emoji, display)
		}

		if *fix {
			// f.Path is relative to cfg.Root with forward slashes;
			// join it back to the root for filesystem operations.
			absPath := filepath.Join(*root, filepath.FromSlash(f.Path))
			var n int
			var werr error
			if *sub {
				n, werr = demojify.ReplaceFile(absPath, repl)
			} else {
				var changed bool
				changed, werr = demojify.SanitizeFile(absPath, cfg.Options)
				if changed {
					n = len(f.Matches)
				}
			}
			if werr != nil {
				fmt.Fprintf(os.Stderr, "  [FAIL] write %s: %v\n", f.Path, werr)
				exitCode = 1
			} else {
				fmt.Printf("  [PASS] fixed %d occurrence(s)\n", n)
			}
		} else {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
