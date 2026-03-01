# demojify-sanitize

[![Go Reference](https://pkg.go.dev/badge/github.com/nicholashoule/demojify-sanitize.svg)](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize)

A dependency-free Go module that helps developers of web applications and APIs
audit, detect, and fix emoji clutter and redundant whitespace in text content
before it reaches production. Run it as a post-processing step after AI agent
output, as a content gate in your request pipeline, or as a CI quality gate --
one call to `Sanitize` finds and fixes every issue in one pass.

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

Clean LLM output before storing or displaying it:

```go
aiResponse := "Certainly! Here is a summary.\n\n" +
    "The deployment pipeline \U0001F680 runs on every push to main.\n" +
    "Check the dashboard \U0001F4CA for metrics."

clean := demojify.Sanitize(aiResponse, demojify.DefaultOptions())
// "Certainly! Here is a summary.\n\nThe deployment pipeline runs on every push to main.\nCheck the dashboard for metrics."
```

### Content gate — reject or flag emoji-bearing input

```go
if demojify.ContainsEmoji(userInput) {
 opts := demojify.Options{RemoveEmojis: true, NormalizeWhitespace: true}
 userInput = demojify.Sanitize(userInput, opts)
}
```

### HTTP handler — sanitize request body

```go
http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
 body, _ := io.ReadAll(r.Body)
 clean := demojify.Sanitize(string(body), demojify.DefaultOptions())
 // use clean for storage or further processing
})
```

### Pre-commit / CI -- sanitize Markdown files in place

```go
opts := demojify.DefaultOptions()
for _, path := range markdownPaths {
 data, _ := os.ReadFile(path)
 _ = os.WriteFile(path, []byte(demojify.Sanitize(string(data), opts)), 0o644)
}
```

### Directory scanner -- audit a repo in one call

```go
cfg := demojify.DefaultScanConfig()
cfg.SkipDirs = append(cfg.SkipDirs, "docs/", "third_party/")

findings, err := demojify.ScanDir(cfg)
if err != nil {
 log.Fatal(err)
}
for _, f := range findings {
 fmt.Printf("%s: has_emoji=%v\n", f.Path, f.HasEmoji)
 // Auto-fix: os.WriteFile(f.Path, []byte(f.Cleaned), 0o644)
}
```

### Substitution -- replace emoji with meaningful text

Use `DefaultReplacements()` to map emoji to text equivalents, then write
changed files back atomically:

```go
// One-shot string substitution.
repl := demojify.DefaultReplacements()
clean := demojify.Replace("\u2705 tests passed, \u274c build failed", repl)
// "[PASS] tests passed, [FAIL] build failed"

// Rewrite a file in place (no write when nothing changes).
n, err := demojify.ReplaceFile("output.log", repl)
fmt.Printf("replaced %d occurrence(s)\n", n)

// Scan a directory, then substitute each dirty file.
cfg := demojify.DefaultScanConfig()
cfg.Root = "."
cfg.Replacements = demojify.DefaultReplacements()
findings, _ := demojify.ScanDir(cfg)
for _, f := range findings {
 demojify.ReplaceFile(f.Path, repl)
}
```

Callers needing per-occurrence detail (line, column, surrounding context)
can enable `CollectMatches` on the scan config, or call `FindMatchesInFile`
directly:

```go
matches, err := demojify.FindMatchesInFile("output.log", demojify.DefaultReplacements())
for _, m := range matches {
 fmt.Printf("line %d col %d: %q -> %q\n", m.Line, m.Column, m.Emoji, m.Replacement)
}
```

## API

### Substitution functions

#### `Replace(text string, replacements map[string]string) string`

Substitutes emoji codepoints using the provided map (longest key wins, so
variation-selector sequences like `"\u26a0\ufe0f"` are matched before their bare
codepoint), then strips any remaining unmatched emoji via `Demojify`.
A nil or empty map behaves identically to `Demojify`.

```go
repl := demojify.DefaultReplacements()
demojify.Replace("\u2705 OK \u274c FAIL", repl)
// "[PASS] OK [FAIL] FAIL"
```

#### `ReplaceFile(path string, replacements map[string]string) (count int, err error)`

Reads `path`, applies `Replace`, and atomically writes the result back only
if changes were made. Original file permissions are preserved. Returns 0 and
no write when the file is already clean.

#### `ReplaceCount(text string, replacements map[string]string) (string, int)`

Applies `Replace` and returns both the cleaned string and the total number of
substitutions and removals performed. Equivalent to calling `Replace` and
then counting separately, but in a single pass.

#### `FindAll(text string) []string`

Returns the distinct emoji codepoint sequences found in `text` in
first-occurrence order. Each sequence appears at most once. Uses `emojiRE`
(not a replacement map), so it finds every emoji regardless of mapping.

#### `FindAllMapped(text string, replacements map[string]string) []string`

Returns only the distinct keys from `replacements` that appear in `text`,
in first-occurrence order. Uses the same longest-first greedy walk as
`Replace`, so `"\u26a0\ufe0f"` wins over bare `"\u26a0"` when both are in the map.

```go
repl := demojify.DefaultReplacements()
demojify.FindAllMapped("\u2705 pass \u274c fail", repl)
// ["\u2705", "\u274c"]
```

#### `FindMatchesInFile(path string, replacements map[string]string) ([]Match, error)`

Reads `path` and returns a `Match` for every emoji codepoint occurrence,
with `Replacement` populated from the map. Returns nil (no error) when the
file contains no emoji.

#### `DefaultReplacements() map[string]string`

Returns a fresh copy of the built-in ~100-entry emoji-to-text map. Callers
may add, remove, or override entries without affecting other callers.

Categories covered:

| Category | Examples |
|---|---|
| Warning / alerts | U+26A0 (warning sign) -> `WARNING`, U+203C (double exclamation) -> `[ALERT]` |
| Status symbols | U+2705 (check mark) -> `[PASS]`, U+274C (cross mark) -> `[FAIL]`, U+2757 -> `[ALERT]` |
| Favorites / annotations | U+2B50 (star) -> `[FEATURED]`, U+1F4A1 (light bulb) -> `[TIP]`, U+1F4CC (pushpin) -> `[PINNED]` |
| Tech / deployment | U+1F4BB (laptop) -> `Code`, U+1F5A5 (computer) -> `Server`, U+2699 (gear) -> `Configuration` |
| Arrows | U+2192 (right arrow) -> `->`, U+2190 (left arrow) -> `<-`, U+21D2 (double right) -> `=>` |
| Geometric shapes | U+25CF (black circle) -> `*`, U+25CB (white circle) -> `o`, U+25B2 (up triangle) -> `^` |
| Checkboxes | U+2610 (ballot box) -> `[ ]`, U+2611 (checked box) -> `[x]`, U+2612 (X box) -> `[x]` |
| Dingbats | U+2022 (bullet) -> `*`, U+2764 (heart) -> `<3`, U+2666 (diamond suit) -> `<>` |

#### `Match` struct

Returned by `FindMatchesInFile` and populated in `Finding.Matches` when
`ScanConfig.CollectMatches` is true:

```go
type Match struct {
 Emoji string // matched codepoint sequence
 Replacement string // value from replacements map; empty if not mapped
 Line int // 1-based line number
 Column int // 0-based byte offset within the line
 Context string // full line text
}
```

### `Demojify(text string) string`

Removes every emoji and Unicode pictographic character from `text`. All
surrounding ASCII and non-emoji Unicode text (e.g. Chinese, Arabic) is left
unchanged.

```go
demojify.Demojify("\U0001F680 Deploy complete! \U0001F4CA")
// -> " Deploy complete! "
```

### `ContainsEmoji(text string) bool`

Reports whether `text` contains at least one emoji or Unicode pictographic
character recognised by `Demojify`.

```go
demojify.ContainsEmoji("Hello \U0001F600")  // -> true
demojify.ContainsEmoji("Hello")              // -> false
```

### `Normalize(text string) string`

Collapses redundant whitespace:

- consecutive spaces/tabs → single space
- trailing whitespace before a newline → removed
- three or more consecutive blank lines → two blank lines
- leading/trailing whitespace of the whole string → trimmed

```go
demojify.Normalize("Hello World\n\n\n\nMore text")
// → "Hello World\n\nMore text"
```

### `Sanitize(text string, opts Options) string`

Applies a configurable pipeline in order: emoji removal → whitespace
normalization. Each step is independently toggled through `Options`.

```go
// All steps on (recommended).
clean := demojify.Sanitize(text, demojify.DefaultOptions())

// Only strip emojis; leave everything else untouched.
clean := demojify.Sanitize(text, demojify.Options{RemoveEmojis: true})
```

### `Options` / `DefaultOptions()`

```go
type Options struct {
 RemoveEmojis        bool                  // strip emoji / pictographic characters
 NormalizeWhitespace bool                  // collapse redundant spaces and blank lines
 AllowedRanges       []*unicode.RangeTable // emoji codepoints to preserve during removal
}

func DefaultOptions() Options // all fields true; AllowedRanges nil (no exceptions)
```

`AllowedRanges` lets callers preserve specific emoji codepoints while still
removing all others. A codepoint is kept when it belongs to any table in the
slice. `nil` (the default) removes every matched codepoint.

```go
// Remove all emoji except the rocket (U+1F680).
clean := demojify.Sanitize(text, demojify.Options{
 RemoveEmojis: true,
 AllowedRanges: []*unicode.RangeTable{
 {R32: []unicode.Range32{{Lo: 0x1F680, Hi: 0x1F680, Stride: 1}}},
 },
})
```

### `ScanConfig` / `DefaultScanConfig()` / `ScanDir` / `ScanFile`

The scanner walks a directory tree and returns a `Finding` for every file whose
content would change after sanitization. Configure exemptions through
`ScanConfig`:

```go
type ScanConfig struct {
 Root string // directory to scan ("." if empty)
 SkipDirs []string // directory prefixes to skip (e.g., ".git/", "vendor/")
 ExemptFiles []string // base filenames to skip (e.g., "README.md")
 ExemptSuffixes []string // file suffixes to skip (e.g., "_test.go")
 Extensions []string // file types to scan; nil (default) = all files
 Options Options // sanitization pipeline applied to each file
 Replacements map[string]string // if set, ScanDir uses Replace instead of Sanitize
 CollectMatches bool // if true, populate Finding.Matches for each finding
}

func DefaultScanConfig() ScanConfig
// Root: ".", SkipDirs: [".git/", "vendor/", "node_modules/"],
// ExemptSuffixes: ["_test.go"], Extensions: nil (all file types),
// Options: RemoveEmojis only
```

By default all file types are scanned (`.go`, `.md`, `.txt`, `.yaml`, `.ini`,
`.csv`, `.py`, `.rs`, etc.). To restrict to specific extensions:

```go
cfg := demojify.DefaultScanConfig()
cfg.Extensions = []string{".go", ".md"} // scan only Go and Markdown
```

`ScanDir` walks the full tree; `ScanFile` checks a single file:

```go
findings, err := demojify.ScanDir(cfg) // []Finding, error
finding, err := demojify.ScanFile(path, opts) // *Finding, error
```

Each `Finding` contains the file path, whether emoji was detected, the
original content, and the cleaned content ready to write back:

```go
type Finding struct {
 Path string // forward-slash normalized path
 HasEmoji bool // ContainsEmoji detected emoji
 Original string // content before sanitization
 Cleaned string // content after Sanitize or Replace
 Matches []Match // populated when ScanConfig.CollectMatches is true
}
```

## Unicode emoji coverage

The following Unicode blocks are stripped by `Demojify`:

| Range | Block |
|---|---|
| U+2600–U+27BF | Miscellaneous Symbols + Dingbats |
| U+231A–U+23FA | Clocks, media controls |
| U+25AA–U+25FE | Geometric shapes used as emoji |
| U+2934–U+2935, U+2B05–U+2B07 | Curved / directional arrows |
| U+2B1B–U+2B55 | Large squares, star, circle |
| U+3030, U+303D, U+3297, U+3299 | CJK/wavy dash symbols |
| U+1F000–U+1FAFF | All supplementary emoji blocks |
| U+200D | Zero Width Joiner |
| U+20E3 | Combining Enclosing Keycap |
| U+FE00–U+FE0F | Variation Selectors 1–16 |

Intentionally **not** removed: `©` (U+00A9), `®` (U+00AE), `™` (U+2122),
basic mathematical/technical arrows (U+2190–U+2193), and all non-emoji Unicode
scripts.

## Design

See [docs/design.md](docs/design.md) for the rationale behind key technical
decisions (zero-dependency policy, pipeline order, false-positive prevention,
and intentional Unicode exclusions).

## License

See [LICENSE](LICENSE).