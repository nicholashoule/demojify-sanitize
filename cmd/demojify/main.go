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
//	-normalize       collapse redundant whitespace in all scanned files;
//	                 implies -fix
//	-quiet           suppress all output; exit code only (0 = clean, 1 = findings/errors)
//	-exts <.go,.md>  comma-separated extensions to scan (default: all files);
//	                 a leading dot is added automatically if omitted
//	-version         print version and exit
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func main() {
	root := flag.String("root", ".", "directory to scan")
	fix := flag.Bool("fix", false, "rewrite affected files in place")
	sub := flag.Bool("sub", false, "substitute emoji with text tokens (implies -fix)")
	normalize := flag.Bool("normalize", false, "collapse redundant whitespace in all scanned files (implies -fix)")
	quiet := flag.Bool("quiet", false, "suppress all output; exit code only (0 = clean, 1 = findings/errors)")
	exts := flag.String("exts", "", "comma-separated extensions to scan, e.g. .go,.md (default: all)")
	version := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *version {
		fmt.Println(cliVersion())
		return
	}

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
		if !*quiet {
			fmt.Println("[PASS] no emoji found")
		}
		return
	}

	exitCode := 0
	for _, f := range findings {
		if !*quiet {
			fmt.Printf("\n[WARN] %s\n", f.Path)
			if len(f.Matches) == 0 && !f.HasEmoji {
				// File changed due to whitespace normalization only.
				fmt.Println("  (whitespace normalized, no emoji found)")
			}
			for _, m := range f.Matches {
				display := m.Replacement
				if display == "" {
					display = "(stripped)"
				}
				fmt.Printf("  line %d col %d: %q -> %q\n", m.Line, m.Column, m.Sequence, display)
			}
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
			} else if !*quiet {
				if n == 0 && !f.HasEmoji {
					fmt.Printf("  [PASS] fixed 1 file (whitespace only)\n")
				} else {
					fmt.Printf("  [PASS] fixed %d occurrence(s)\n", n)
				}
			}
		} else {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

// cliVersion returns the module version reported by the Go build system.
// A semver tag (e.g. "v0.2.1") is embedded only when the binary is installed
// from a published tagged release (e.g. "go install ...@v0.2.1"). Builds from
// local source -- whether via "go run", "go build", or "go install" without a
// version suffix -- have the version set to "(devel)" by the Go toolchain.
// The empty-string fallback is a defensive guard for unusual non-module build
// contexts where debug.ReadBuildInfo() succeeds but Main.Version is unset.
func cliVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "demojify version unknown"
	}
	v := info.Main.Version
	if v == "" {
		v = "(devel)"
	}
	return "demojify " + v
}
