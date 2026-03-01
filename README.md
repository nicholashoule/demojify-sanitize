# demojify-sanitize

[![Go Reference](https://pkg.go.dev/badge/github.com/nicholashoule/demojify-sanitize.svg)](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize)

A dependency-free Go module that helps developers of web applications and APIs
audit, detect, and fix emoji clutter, AI-generated preamble phrases, and
redundant whitespace before they reach production. Run it as a post-processing
step after AI agent output, as a content gate in your request pipeline, or as a
CI check -- one call to `Sanitize` finds and fixes every issue in one pass.

## Install

```bash
go get github.com/nicholashoule/demojify-sanitize
```

## Quick start

```go
import demojify "github.com/nicholashoule/demojify-sanitize"

// Remove all emojis, AI preamble phrases, and redundant whitespace in one call.
clean := demojify.Sanitize(text, demojify.DefaultOptions())
```

## Integration patterns

### AI response post-processing

Clean LLM output before storing or displaying it:

```go
aiResponse := "Certainly! Here is a summary.\n\n" +
    "The deployment pipeline \U0001F680 runs on every push to main.\n" +
    "Check the dashboard \U0001F4CA for metrics."

clean := demojify.Sanitize(aiResponse, demojify.DefaultOptions())
// "Here is a summary.\n\nThe deployment pipeline  runs on every push to main.\nCheck the dashboard  for metrics."
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

## API

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

Applies a configurable pipeline in order: emoji removal → AI-clutter removal
→ whitespace normalization. Each step is independently toggled through
`Options`.

```go
// All steps on (recommended).
clean := demojify.Sanitize(text, demojify.DefaultOptions())

// Only strip emojis; leave everything else untouched.
clean := demojify.Sanitize(text, demojify.Options{RemoveEmojis: true})
```

### `Options` / `DefaultOptions()`

```go
type Options struct {
 RemoveEmojis bool // strip emoji / pictographic characters
 RemoveAIClutter bool // strip AI preamble and boilerplate phrases
 NormalizeWhitespace bool // collapse redundant spaces and blank lines
 AllowedRanges []*unicode.RangeTable // emoji codepoints to preserve during removal
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
}

func DefaultScanConfig() ScanConfig
// Root: ".", SkipDirs: [".git/", "vendor/", "node_modules/"],
// ExemptSuffixes: ["_test.go"], Extensions: nil (all file types),
// Options: RemoveEmojis + RemoveAIClutter
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
 Cleaned string // content after Sanitize
}
```

**AI-clutter patterns removed** (case-insensitive, must start a line):

| Pattern | Example match |
|---|---|
| `Certainly[!,.]` | `Certainly! Here is…` |
| `Sure[!,.]` | `Sure, ` |
| `Of course[!,.]` | `Of course! ` |
| `Absolutely[!,.]` | `Absolutely, ` |
| `Great[!,.]` | `Great! ` |
| `Excellent[!,.]` | `Excellent. ` |
| `Noted[!,.]` | `Noted. ` |
| `I'd be happy to…` | `I'd be happy to help!` |
| `I can help…` | `I can certainly help with that.` |
| `I'll help you with that` | `I'll help you with that.` |
| `Let me help you` | `Let me help you.` |
| `I hope this helps` | `I hope this helps!` |
| `Feel free to ask…` | `Feel free to ask if you need help.` |

Unambiguous short words (`Sure`, `Great`, etc.) require trailing punctuation
(`!`, `,`, or `.`) to prevent false positives on legitimate text like
`"Sure enough, the build passed."`.

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