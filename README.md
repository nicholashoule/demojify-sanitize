# demojify-sanitize

[![CI](https://github.com/nicholashoule/demojify-sanitize/actions/workflows/ci.yml/badge.svg)](https://github.com/nicholashoule/demojify-sanitize/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nicholashoule/demojify-sanitize.svg)](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nicholashoule/demojify-sanitize)](go.mod)
[![License](https://img.shields.io/github/license/nicholashoule/demojify-sanitize)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-none-brightgreen)](go.mod)

A dependency-free Go library for auditing and detecting emoji clutter, removing or substituting it, and normalizing unnecessary whitespace in text content. It is built for post-processing AI agent output, sanitizing AI or user-submitted content, and functioning as a CI quality gate. `Sanitize` runs removal and normalization as one pipeline, `Replace` substitutes mapped emoji with text equivalents and strips the rest, `ScanDir` audits directory trees, and `ContainsEmoji` provides direct single-string detection.

## Features

- **Emoji removal** -- strips emoji-related codepoints in the documented Unicode ranges; ZWJ, variation-selector, keycap, and tag codepoints are handled
- **Whitespace normalization** -- collapses redundant inline spaces and blank lines while preserving leading indentation
- **Configurable pipeline** -- `Sanitize` runs removal and normalization in one call; `AllowedRanges` and `AllowedEmojis` let callers preserve specific codepoints
- **Safe by construction** -- `SanitizeSafe`, `SafeOptions` and `SafeScanConfig` (v1.2.0) always remove emoji and keep script joiners, so no unset option can turn sanitizing off
- **Script-aware** -- `ScriptAware` (v1.1.0; on in `SafeOptions`) keeps the Zero Width Joiner and standardized variation selectors in Indic, CJK and mathematical text while still removing every emoji; `ContainsEmojiWith` detects under the same options
- **Substitution** -- `Replace` maps 280 built-in emoji to readable text equivalents (e.g., `[PASS]`, `[FAIL]`); custom maps supported
- **Metrics** -- `SanitizeReport` returns emoji count removed and bytes saved alongside the cleaned text
- **Streaming** -- `SanitizeReader` processes `io.Reader` line by line; supports lines up to 1 MiB
- **JSON-aware** -- `SanitizeJSON` cleans string values only; preserves keys, numbers, booleans, null, and numeric precision
- **Directory scanner** -- `ScanDir` / `ScanDirContext` walk an entire tree and return per-file findings; cancellation supported
- **Safer writes** -- `SanitizeFile`, `ReplaceFile`, `WriteFinding`, and `FixDir` use a same-directory temp file and rename while preserving permissions, and never overwrite a file changed since it was read (`ErrFileChanged`); replacement is atomic on POSIX and best-effort on Windows
- **CLI** -- `cmd/demojify` supports audit, strip (`-fix`), substitute (`-sub`), normalize, quiet mode, extension filter, and directory skip
- **Zero external dependencies** -- pure stdlib; no `go.sum` required

## Installation

```bash
go get github.com/nicholashoule/demojify-sanitize@v1.1.0
```

### CLI

```bash
go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@v1.1.0
```

```bash
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v1.1.0 -sub
```

### Quick start

The v1 module keeps the existing import path; no `/v1` suffix is required.

```go
import demojify "github.com/nicholashoule/demojify-sanitize"

// Recommended (v1.2.0): always removes emoji, keeps script joiners,
// leaves whitespace alone. No Options value to leave unset by mistake.
clean := demojify.SanitizeSafe(text)

// Also collapse redundant whitespace:
opts := demojify.SafeOptions()
opts.NormalizeWhitespace = true
clean = demojify.Sanitize(text, opts)
```

`SafeOptions`, `SanitizeSafe` and `SafeScanConfig` are the recommended path.
The v1 defaults stay as they were (semantic versioning): a zero `Options`
passes text through unchanged, and `DefaultOptions`, `DefaultScanConfig`
and `ContainsEmoji` are not script-aware. They are stable, not deprecated,
but easy to misuse; see [Options](#options).

A ready-to-run CLI example lives in [cmd/demojify/main.go](cmd/demojify/main.go).
It audits a directory tree for emoji, reports every occurrence with file, line,
and column, and optionally rewrites affected files (`-fix`) or substitutes emoji
with text tokens (`-sub`). Use `-skip` to exclude specific directories
(e.g., `dist`, `build`) in addition to the defaults.

```bash
# Build once, then run the binary:
go build -o demojify ./cmd/demojify
./demojify -root . -sub -skip dist,build
```

## Integration patterns

### AI response post-processing

```go
clean := demojify.SanitizeSafe(aiResponse)
```

### Content gate -- detect and clean

```go
if demojify.ContainsEmojiWith(userInput, demojify.SafeOptions()) {
	userInput = demojify.SanitizeSafe(userInput)
}
```

### Directory scanner -- audit a repo in one call

```go
cfg := demojify.SafeScanConfig()
findings, _ := demojify.ScanDir(cfg)
for _, f := range findings {
	fmt.Printf("%s: has_emoji=%v\n", f.Path, f.HasEmoji)
}
```

### Batch fix -- scan and write back in one call

```go
cfg := demojify.SafeScanConfig()
fixed, _, err := demojify.FixDir(".", cfg)
fmt.Printf("fixed %d file(s)\n", fixed)
```

A file edited after the scan read it is left alone: the error wraps
`ErrFileChanged` (from v1.2.0; `WriteFinding`, `SanitizeFile` and
`ReplaceFile` check the same way).

### Substitution -- replace emoji with meaningful text

```go
repl := demojify.DefaultReplacements()
clean := demojify.Replace("\u2705 tests passed, \u274c build failed", repl)
// "[PASS] tests passed, [FAIL] build failed"
```

### Git pre-commit hook

Run the CLI from a hook to block emoji at commit time. [docs/git-hooks.md](docs/git-hooks.md)
has the pre-built-binary and `go run` (with repogov) hooks, auto-fix,
substitution, the Go API variant, and macOS/Linux/Windows examples.

### Streaming sanitization

Process LLM token streams or HTTP chunked responses line by line without
buffering the full input:

```go
var out bytes.Buffer
err := demojify.SanitizeReader(llmStream, &out, demojify.SafeOptions())
```

Lines up to 1 MiB are supported. Longer lines return `bufio.ErrTooLong`.

### JSON value sanitization

Clean string values inside a JSON document while leaving keys, numbers,
booleans, and null untouched. Never run `Sanitize` over raw JSON bytes:
it would edit text inside keys and strings without knowing the structure.

```go
clean, err := demojify.SanitizeJSON(jsonBytes, demojify.SafeOptions())
```

Returns an error for invalid JSON and for input with trailing non-whitespace
content after the first value (e.g., `{"a":1} trailing`). A second complete
JSON value returns `ErrMultipleJSONValues`.

See [example_test.go](example_test.go) for additional runnable patterns
(HTTP handler, pre-commit/CI, file write-back, per-occurrence matching).

## API

Full signatures and doc comments are on
[pkg.go.dev](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize).

### Core functions

| Function | Purpose |
|----------|---------|
| `SanitizeSafe(text) string` | Recommended: `Sanitize` with `SafeOptions()` (v1.2.0) |
| `Sanitize(text, opts) string` | Configurable pipeline: emoji removal then whitespace normalization |
| `SanitizeFile(path, opts) (bool, error)` | Sanitize through temp-file replacement; no write when clean, `ErrFileChanged` when edited since the read |
| `Demojify(text) string` | Strip recognized emoji-related codepoints |
| `ContainsEmoji(text) bool` | Detect emoji presence |
| `ContainsEmojiWith(text, opts) bool` | Detect what `Sanitize` with opts would remove (honors `ScriptAware`, `AllowedRanges`, `AllowedEmojis`); a zero `Options` answers as `ContainsEmoji` |
| `CountEmoji(text) int` | Count emoji codepoint occurrences |
| `BytesSaved(text) int` | Bytes freed by emoji removal |
| `Normalize(text) string` | Collapse redundant whitespace (preserves leading indentation) |
| `TechnicalSymbolRanges() []*unicode.RangeTable` | Pre-built ranges for check marks, gears, etc. -- pass to `AllowedRanges` |

### Reporting and streaming

| Function / Type | Purpose |
|-----------------|--------|
| `SanitizeReport(text, opts) SanitizeResult` | Sanitize with structured metrics (emoji count, bytes saved) |
| `SanitizeResult` | Cleaned text plus `EmojiRemoved` and `BytesSaved` fields |
| `SanitizeReader(r, w, opts) error` | Line-by-line streaming sanitization (LLM streams, MCP payloads) |
| `SanitizeJSON(data, opts) ([]byte, error)` | Sanitize JSON string values only; preserves structure and numeric precision |

### Substitution

| Function | Purpose |
|----------|---------|
| `Replace(text, repl) string` | Map emoji to text equivalents; strip unmapped remainder |
| `ReplaceFile(path, repl) (int, error)` | Temp-file replacement; no write when clean |
| `ReplaceCount(text, repl) (string, int)` | Replace and return substitution count |
| `FindAll(text) []string` | Distinct emoji sequences in text |
| `FindAllMapped(text, repl) []string` | Mapped keys found in text |
| `DefaultReplacements() map[string]string` | Built-in 280-entry emoji-to-text map ([full list](docs/replacements.md)) |

### Scanner

| Function / Type | Purpose |
|-----------------|---------|
| `ScanDir(cfg) ([]Finding, error)` | Walk directory tree, return findings |
| `ScanDirContext(ctx, cfg) ([]Finding, error)` | Context-aware scan with cancellation support |
| `ScanFile(path, opts) (*Finding, error)` | Check a single file |
| `FindMatchesInFile(path, repl) ([]Match, error)` | Per-occurrence match detail (line, column, context) |
| `WriteFinding(path, f) (bool, error)` | Temp-file write-back; `ErrFileChanged` (no write) when the file changed since the scan |
| `FixDir(root, cfg) (fixed, clean int, err error)` | Scan and fix an entire directory tree in one call |
| `SafeScanConfig()` | Recommended: `DefaultScanConfig()` with `SafeOptions()` (v1.2.0) |
| `ScanConfig` / `DefaultScanConfig()` | Scanner configuration (root, skip dirs, extensions, `MaxFileBytes` = `DefaultMaxFileBytes`, etc.) |
| `Finding` | Path, HasEmoji, Original, Cleaned, Matches |
| `Match` | Sequence, Replacement, Line, Column, Context |

### Options

```go
type Options struct {
	RemoveEmojis        bool               // strip emoji / pictographic characters
	NormalizeWhitespace bool               // collapse redundant spaces and blank lines
	AllowedRanges       []*unicode.RangeTable // preserve emoji in these Unicode ranges
	AllowedEmojis       []string           // preserve specific emoji strings (exact match)
	ScriptAware         bool               // keep ZWJ and FE00-FE0D outside emoji sequences (v1.1.0)
}

func SafeOptions() Options    // recommended: RemoveEmojis + ScriptAware (v1.2.0)
func DefaultOptions() Options // RemoveEmojis + NormalizeWhitespace (not script-aware)
```

The zero `Options` disables every step, so `Sanitize(text, Options{})` returns
`text` unchanged: stable v1 behavior that a forgotten field turns into a silent
pass-through. Start from `SafeOptions()` instead.

`AllowedRanges` and `AllowedEmojis` can be combined. Empty strings in
`AllowedEmojis` and empty keys in replacement maps are silently skipped.

`ScriptAware` is off in a zero `Options` and in `DefaultOptions`, so v1
callers see no change; `SafeOptions` turns it on. Without it, the Zero Width
Joiner that Devanagari, Malayalam and Sinhala use is removed (and
`ContainsEmoji` reports it) wherever it appears.

```go
// An input gate that accepts Indic text but refuses an emoji.
if demojify.ContainsEmojiWith(userInput, demojify.SafeOptions()) {
	return errEmojiNotAllowed
}
```

```go
// Remove recognized emoji except rocket and thumbs-up.
opts := demojify.SafeOptions()
opts.AllowedEmojis = []string{"\U0001F680", "\U0001F44D"}
clean := demojify.Sanitize(text, opts)
```

## Unicode emoji coverage

`Demojify` strips U+2139, U+2600-U+27BF, U+1F000-U+1FAFF, ZWJ (U+200D),
variation selectors (U+FE00-U+FE0F), tag characters (U+E0020-U+E007F), and
related auxiliary ranges. Intentionally **not** removed: copyright, registered,
trademark, and basic math/technical arrows. With `ScriptAware`, the joiner and
the standardized variation selectors (U+FE00-U+FE0D) are removed only inside
emoji sequences.

Full range table: [docs/unicode-coverage.md](docs/unicode-coverage.md).

## Design and documentation

| Document | Contents |
|----------|----------|
| [docs/design.md](docs/design.md) | Architecture rationale: zero-dependency policy, pipeline order, error handling, temp-file write strategy |
| [docs/replacements.md](docs/replacements.md) | Full `DefaultReplacements()` reference: all 280 entries organized by category |
| [docs/unicode-coverage.md](docs/unicode-coverage.md) | `emojiRE` ranges, intentional exclusions (copyright, trademark, math arrows), substitution vs. stripping |
| [docs/cli.md](docs/cli.md) | `cmd/demojify` CLI reference: flags, exit codes, output format, examples |
| [docs/ci.md](docs/ci.md) | CI pipeline integration: GitHub Actions, GitLab CI, test-based gates, JSON tooling |
| [docs/git-hooks.md](docs/git-hooks.md) | Pre-commit hook integration: shell and Go examples, auto-fix, substitution |

## License

See [LICENSE](LICENSE).
