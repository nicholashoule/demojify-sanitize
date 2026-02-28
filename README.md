# demojify-sanitize

[![Go Reference](https://pkg.go.dev/badge/github.com/nicholashoule/demojify-sanitize.svg)](https://pkg.go.dev/github.com/nicholashoule/demojify-sanitize)

A dependency-free Go module for detecting and removing emojis, Unicode
pictographic characters, AI-generated clutter, and other non-semantic
artifacts from Markdown, documentation, and repository text files.

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

## API

### `Demojify(text string) string`

Removes every emoji and Unicode pictographic character from `text`. All
surrounding ASCII and non-emoji Unicode text (e.g. Chinese, Arabic) is left
unchanged.

```go
demojify.Demojify("🚀 Deploy complete! 📊")
// → " Deploy complete! "
```

### `ContainsEmoji(text string) bool`

Reports whether `text` contains at least one emoji or Unicode pictographic
character recognised by `Demojify`.

```go
demojify.ContainsEmoji("Hello 😀")  // → true
demojify.ContainsEmoji("Hello")     // → false
```

### `Normalize(text string) string`

Collapses redundant whitespace:

- consecutive spaces/tabs → single space
- trailing whitespace before a newline → removed
- three or more consecutive blank lines → two blank lines
- leading/trailing whitespace of the whole string → trimmed

```go
demojify.Normalize("Hello   World\n\n\n\nMore text")
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
    RemoveEmojis        bool // strip emoji / pictographic characters
    RemoveAIClutter     bool // strip AI preamble and boilerplate phrases
    NormalizeWhitespace bool // collapse redundant spaces and blank lines
}

func DefaultOptions() Options // all fields true
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

## License

See [LICENSE](LICENSE).

