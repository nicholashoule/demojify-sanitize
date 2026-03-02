# demojify-sanitize

[![CI](https://github.com/nicholashoule/demojify-sanitize/actions/workflows/ci.yml/badge.svg)](https://github.com/nicholashoule/demojify-sanitize/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nicholashoule/demojify-sanitize.svg)](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nicholashoule/demojify-sanitize)](go.mod)
[![License](https://img.shields.io/github/license/nicholashoule/demojify-sanitize)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-none-brightgreen)](go.mod)

A dependency-free Go module for auditing, detecting, removing, and substituting emoji clutter and redundant whitespace in text content before it reaches production. Use it as a post-processing step after AI agent output, as a content gate in your request pipeline, or as a CI quality gate -- one call to `Sanitize` strips and normalizes in a single pass, `Replace` maps emoji to meaningful text equivalents, `ScanDir` audits entire directory trees (it calls `ContainsEmoji` internally per file), and `ContainsEmoji` is available directly for ad-hoc single-string detection.

## Install

```bash
go get github.com/nicholashoule/demojify-sanitize
```

## Quick start

```go
import demojify "github.com/nicholashoule/demojify-sanitize"

// Remove all emojis and normalize whitespace in one call.
clean := demojify.Sanitize(text, demojify.DefaultOptions())
```

A ready-to-run CLI example lives in [cmd/demojify/main.go](cmd/demojify/main.go).
It audits a directory tree for emoji, reports every occurrence with file, line,
and column, and optionally rewrites affected files (`-fix`) or substitutes emoji
with text tokens (`-sub`).

```bash
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify -root . -sub
```

## Integration patterns

### AI response post-processing

```go
clean := demojify.Sanitize(aiResponse, demojify.DefaultOptions())
```

### Content gate -- detect and clean

```go
if demojify.ContainsEmoji(userInput) {
 userInput = demojify.Sanitize(userInput, demojify.DefaultOptions())
}
```

### Directory scanner -- audit a repo in one call

```go
cfg := demojify.DefaultScanConfig()
findings, _ := demojify.ScanDir(cfg)
for _, f := range findings {
 fmt.Printf("%s: has_emoji=%v\n", f.Path, f.HasEmoji)
}
```

### Substitution -- replace emoji with meaningful text

```go
repl := demojify.DefaultReplacements()
clean := demojify.Replace("\u2705 tests passed, \u274c build failed", repl)
// "[PASS] tests passed, [FAIL] build failed"
```

See [example_test.go](example_test.go) for additional runnable patterns
(HTTP handler, pre-commit/CI, file write-back, per-occurrence matching).

## API

Full signatures and doc comments are on
[pkg.go.dev](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize).

### Core functions

| Function | Purpose |
|----------|---------|
| `Sanitize(text, opts) string` | Configurable pipeline: emoji removal then whitespace normalization |
| `SanitizeFile(path, opts) (bool, error)` | Sanitize a file atomically; no write when clean |
| `Demojify(text) string` | Strip all emoji / pictographic codepoints |
| `ContainsEmoji(text) bool` | Detect emoji presence |
| `Normalize(text) string` | Collapse redundant whitespace (preserves leading indentation) |

### Substitution

| Function | Purpose |
|----------|---------|
| `Replace(text, repl) string` | Map emoji to text equivalents; strip unmapped remainder |
| `ReplaceFile(path, repl) (int, error)` | Atomic in-place replacement; no write when clean |
| `ReplaceCount(text, repl) (string, int)` | Replace and return substitution count |
| `FindAll(text) []string` | Distinct emoji sequences in text |
| `FindAllMapped(text, repl) []string` | Mapped keys found in text |
| `DefaultReplacements() map[string]string` | Built-in ~137-entry emoji-to-text map ([full list](docs/replacements.md)) |

### Scanner

| Function / Type | Purpose |
|-----------------|---------|
| `ScanDir(cfg) ([]Finding, error)` | Walk directory tree, return findings |
| `ScanFile(path, opts) (*Finding, error)` | Check a single file |
| `FindMatchesInFile(path, repl) ([]Match, error)` | Per-occurrence match detail (line, column, context) |
| `WriteFinding(path, f) (bool, error)` | Atomic write-back without re-reading |
| `ScanConfig` / `DefaultScanConfig()` | Scanner configuration (root, skip dirs, extensions, etc.) |
| `Finding` | Path, HasEmoji, Original, Cleaned, Matches |
| `Match` | Sequence, Replacement, Line, Column, Context |

### Options

```go
type Options struct {
 RemoveEmojis        bool               // strip emoji / pictographic characters
 NormalizeWhitespace bool               // collapse redundant spaces and blank lines
 AllowedRanges       []*unicode.RangeTable // preserve emoji in these Unicode ranges
 AllowedEmojis       []string           // preserve specific emoji strings (exact match)
}

func DefaultOptions() Options // RemoveEmojis + NormalizeWhitespace = true
```

`AllowedRanges` and `AllowedEmojis` can be combined. Empty strings in
`AllowedEmojis` and empty keys in replacement maps are silently skipped.

```go
// Remove all emoji except rocket and thumbs-up.
clean := demojify.Sanitize(text, demojify.Options{
 RemoveEmojis:  true,
 AllowedEmojis: []string{"\U0001F680", "\U0001F44D"},
})
```

## Unicode emoji coverage

`Demojify` strips U+2139, U+2600-U+27BF, U+1F000-U+1FAFF, ZWJ (U+200D),
variation selectors (U+FE00-U+FE0F), tag characters (U+E0020-U+E007F), and
related auxiliary ranges. Intentionally **not** removed: copyright, registered,
trademark, and basic math/technical arrows.

Full range table: [docs/unicode-coverage.md](docs/unicode-coverage.md).

## Design and documentation

| Document | Contents |
|----------|----------|
| [docs/design.md](docs/design.md) | Architecture rationale: zero-dependency policy, pipeline order, error handling, atomic writes |
| [docs/replacements.md](docs/replacements.md) | Full `DefaultReplacements()` reference: all ~137 entries organized by category |
| [docs/unicode-coverage.md](docs/unicode-coverage.md) | `emojiRE` ranges, intentional exclusions (copyright, trademark, math arrows), substitution vs. stripping |
| [docs/cli.md](docs/cli.md) | `cmd/demojify` CLI reference: flags, exit codes, output format, examples |

## License

See [LICENSE](LICENSE).