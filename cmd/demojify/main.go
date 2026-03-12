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
//	-json            output findings as JSON to stdout (overrides -quiet)
//	-exts <.go,.md>  comma-separated extensions to scan (default: all files);
//	                 a leading dot is added automatically if omitted
//	-skip <dirs>     comma-separated directory names to skip in addition to the
//	                 defaults (.git, vendor, node_modules); a trailing slash is
//	                 added automatically if omitted
//	-version         print version and exit
package main

import (
	"encoding/json"
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
	skip := flag.String("skip", "", "comma-separated directory names to skip in addition to defaults, e.g. dist,build")
	jsonOut := flag.Bool("json", false, "output findings as JSON (overrides -quiet)")
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

	if *skip != "" {
		for _, d := range strings.Split(*skip, ",") {
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			// Auto-append trailing slash if missing, matching SkipDirs convention.
			if !strings.HasSuffix(d, "/") {
				d += "/"
			}
			cfg.SkipDirs = append(cfg.SkipDirs, d)
		}
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
		if *jsonOut {
			writeJSON(jsonResult{Findings: []jsonFinding{}})
		} else if !*quiet {
			fmt.Println("[PASS] no emoji found")
		}
		return
	}

	exitCode := 0
	var jFindings []jsonFinding

	for _, f := range findings {
		var jf jsonFinding
		if *jsonOut {
			jf = jsonFinding{
				Path:     f.Path,
				HasEmoji: f.HasEmoji,
			}
			for _, m := range f.Matches {
				jf.Matches = append(jf.Matches, jsonMatch{
					Sequence:    m.Sequence,
					Replacement: m.Replacement,
					Line:        m.Line,
					Column:      m.Column,
					Context:     m.Context,
				})
			}
		} else if !*quiet {
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
			// Always write the fully-cleaned content from the Finding.
			// f.Cleaned has emoji stripped/substituted with inline spaces
			// already collapsed; len(f.Matches) equals the substitution count.
			var changed bool
			changed, werr = demojify.WriteFinding(absPath, f)
			if changed {
				n = len(f.Matches)
			}
			if werr != nil {
				if *jsonOut {
					jf.Fixed = &jsonFix{Error: werr.Error()}
				} else {
					fmt.Fprintf(os.Stderr, "  [FAIL] write %s: %v\n", f.Path, werr)
				}
				exitCode = 1
			} else {
				if *jsonOut {
					jf.Fixed = &jsonFix{Success: true, Count: n}
				} else if !*quiet {
					if n == 0 && !f.HasEmoji {
						fmt.Printf("  [PASS] fixed 1 file (whitespace only)\n")
					} else {
						fmt.Printf("  [PASS] fixed %d occurrence(s)\n", n)
					}
				}
			}
		} else {
			exitCode = 1
		}

		if *jsonOut {
			jFindings = append(jFindings, jf)
		}
	}

	if *jsonOut {
		writeJSON(jsonResult{Findings: jFindings})
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

// jsonResult is the top-level JSON envelope written when -json is set.
type jsonResult struct {
	Findings []jsonFinding `json:"findings"`
}

// jsonFinding describes a single file with findings in JSON output.
type jsonFinding struct {
	Path     string      `json:"path"`
	HasEmoji bool        `json:"hasEmoji"`
	Matches  []jsonMatch `json:"matches,omitempty"`
	Fixed    *jsonFix    `json:"fixed,omitempty"`
}

// jsonMatch describes a single matched codepoint sequence in JSON output.
type jsonMatch struct {
	Sequence    string `json:"sequence"`
	Replacement string `json:"replacement"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Context     string `json:"context"`
}

// jsonFix describes the result of a fix operation on a single file.
type jsonFix struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Error   string `json:"error,omitempty"`
}

// writeJSON encodes v as indented JSON to stdout. If the write fails (e.g.
// broken pipe), a diagnostic is printed to stderr and the process exits with
// code 1.
func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "error writing JSON output:", err)
		os.Exit(1)
	}
}
