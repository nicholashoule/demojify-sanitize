// Command demojify audits a directory tree for emoji, reports every occurrence
// with file, line, and column, and optionally rewrites affected files.
//
// # CLI
//
// Install demojify:
//
//	go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest
//
// Or run without installing:
//
//	go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest [flags]
//
// # CLI Subcommands
//
// demojify has no traditional subcommands. The operational mode is selected
// by flags:
//
//   - Audit (default): scan the directory tree and report all findings.
//   - Fix (-fix): rewrite affected files in place after reporting.
//   - Substitute (-sub): replace emoji with text tokens; implies -fix.
//   - Normalize (-normalize): collapse redundant whitespace; implies -fix.
//
// Modes may be combined; for example, -sub and -normalize apply both
// substitution and whitespace normalization in a single pass.
//
// # CLI Status Markers
//
// Human-readable output uses bracketed status tokens at the start of each line:
//
//   - [PASS]: no emoji found, or all occurrences fixed successfully.
//   - [WARN]: emoji detected in a file; reported with per-occurrence detail.
//   - [FAIL]: a write error occurred while applying a fix.
//
// These markers are suppressed when -quiet is set and replaced by structured
// JSON output when -json is set.
//
// # CLI Flags
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
//
// # CLI Exit Codes
//
//	0  no emoji found, or all findings fixed successfully
//	1  emoji found and -fix not set, a write error occurred, or -root
//	   does not exist / is not a directory
//	2  an unknown flag was passed (flag package parse error)
//
// # CLI JSON Output
//
// With -json, all output is a single JSON object on stdout and the
// human-readable [PASS]/[WARN] text is suppressed. The envelope has one
// key, findings, an array of objects:
//
//   - path: forward-slash relative file path.
//   - hasEmoji: whether the file contains emoji.
//   - matches: per-occurrence detail, omitted when empty; each element has
//     sequence (raw UTF-8 codepoints), replacement (mapped substitute or
//     empty), line (1-based), column (0-based byte offset), and context
//     (the full source line).
//   - fixed: present only with -fix or -sub; has success (bool),
//     count (int), and error (string, omitted on success).
//
// When no findings exist the output is {"findings":[]} with exit code 0.
// The JSON format is a stable, machine-readable API; downstream consumers
// should prefer it over parsing the text format.
//
// # CLI Default Scan Behavior
//
// The CLI uses [demojify.DefaultScanConfig], which skips:
//
//   - Directories: .git/, vendor/, node_modules/ (plus any added via -skip).
//   - Suffixes: *_test.go.
//   - Binary, minified, compressed, and media extensions (e.g. .min.js,
//     .css.map, .gz, .bz2, .zip, .png, .woff2, .pdf, .wasm) -- skipped
//     before the file is opened, so minified assets never produce false
//     positives and the audit performs no I/O on files it cannot act on.
//
// All other file types are scanned unless -exts restricts them. Any
// remaining binary file is detected by a NUL byte in the first 512 bytes
// and skipped, and files larger than 1 MiB are skipped.
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

// usageHeader is printed before the auto-generated flag list. It gives the
// synopsis and the flag-selected operational modes (demojify has no
// subcommands). flag.PrintDefaults appends the per-flag detail after this.
const usageHeader = `demojify -- audit a directory tree for emoji and optionally rewrite affected files.

Usage:
  demojify [flags]

demojify has no subcommands; the operational mode is selected by flags:
  audit (default)   scan and report all findings (exit 1 if any found)
  -fix              rewrite affected files in place after reporting
  -sub              substitute emoji with text tokens (implies -fix)
  -normalize        collapse redundant whitespace (implies -fix)

Flags:
`

// usageFooter is printed after the auto-generated flag list: exit codes and
// a few worked examples, mirroring docs/cli.md.
const usageFooter = `
Exit codes:
  0   no emoji found, or all findings fixed successfully
  1   emoji found without -fix, a write error, or invalid -root
  2   unknown flag (flag parse error)

Examples:
  demojify -root .                 audit the current tree (no writes)
  demojify -root . -fix            strip emoji in place
  demojify -root . -sub            substitute emoji with text tokens
  demojify -root . -exts .go,.md   restrict to Go and Markdown files
  demojify -root . -json           machine-readable JSON output

Full reference: https://github.com/nicholashoule/demojify-sanitize/blob/main/docs/cli.md
`

// usage writes the complete help text. It replaces the flag package default
// (which prints "Usage of <absolute-binary-path>:" with no description,
// synopsis, or examples) so `demojify -h` shows usable documentation.
func usage() {
	fmt.Fprint(os.Stderr, usageHeader)
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, usageFooter)
}

func main() {
	flag.Usage = usage

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
			// already collapsed; len(f.Matches) equals the total number of
			// emoji occurrences (substitutions plus removals).
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
