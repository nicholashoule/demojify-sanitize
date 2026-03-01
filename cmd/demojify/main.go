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
//	-normalize       collapse redundant whitespace left behind by -fix/-sub;
//	                 implies -fix
//	-exts <.go,.md>  comma-separated extensions to scan (default: all files);
//	                 a leading dot is added automatically if omitted
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
	normalize := flag.Bool("normalize", false, "collapse redundant whitespace left behind by -fix/-sub (implies -fix)")
	exts := flag.String("exts", "", "comma-separated extensions to scan, e.g. .go,.md (default: all)")
	flag.Parse()

	if *sub || *normalize {
		*fix = true
	}

	// Validate root exists and is a directory.
	rootInfo, rootErr := os.Stat(*root)
	if rootErr != nil {
		log.Fatalf("root directory: %v", rootErr)
	}
	if !rootInfo.IsDir() {
		log.Fatalf("root is not a directory: %s", *root)
	}

	cfg := demojify.DefaultScanConfig()
	cfg.Root = *root
	cfg.CollectMatches = true
	if *normalize {
		cfg.Options.NormalizeWhitespace = true
	}

	if *exts != "" {
		for _, e := range strings.Split(*exts, ",") {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			// Auto-prepend dot if missing, so "-exts go,md" works like "-exts .go,.md".
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			cfg.Extensions = append(cfg.Extensions, e)
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
			fmt.Printf("  line %d col %d: %q -> %q\n", m.Line, m.Column, m.Sequence, display)
		}

		if *fix {
			// f.Path is relative to cfg.Root with forward slashes;
			// join it back to the root for filesystem operations.
			absPath := filepath.Join(*root, filepath.FromSlash(f.Path))
			var n int
			var werr error
			if *sub && !*normalize {
				// Replacement only: re-apply via ReplaceFile so the count
				// reflects actual substitutions made.
				n, werr = demojify.ReplaceFile(absPath, repl)
			} else {
				// Normalize (with or without sub) or plain fix: write the
				// fully-cleaned content from the Finding in one shot.
				var changed bool
				changed, werr = demojify.WriteFinding(absPath, f)
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
