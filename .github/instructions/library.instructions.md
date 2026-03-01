---
applyTo: "**/*.go"
---

# Library Development Instructions

## Overview

This file defines Go development conventions for the `demojify-sanitize` package.
It applies to all Go source files.

## Key Files

| File | Purpose |
|------|---------|
| [demojify.go](../../demojify.go) | `Demojify`, `ContainsEmoji` -- emoji detection/removal |
| [normalize.go](../../normalize.go) | `Normalize` -- whitespace normalization |
| [sanitize.go](../../sanitize.go) | `Sanitize`, `SanitizeFile`, `Options`, `DefaultOptions` -- pipeline |
| [scan.go](../../scan.go) | `ScanConfig`, `DefaultScanConfig`, `ScanDir`, `ScanFile`, `FindMatchesInFile`, `Finding`, `Match` -- scanner |
| [replace.go](../../replace.go) | `Replace`, `ReplaceFile`, `ReplaceCount`, `FindAll`, `FindAllMapped` -- substitution |
| [replacements.go](../../replacements.go) | `DefaultReplacements` -- built-in emoji-to-text map |
| [write.go](../../write.go) | `WriteFinding` -- atomic write-back for scan results |
| [doc.go](../../doc.go) | Package-level godoc comment |

## Design Principles

**Dependency-free:** Do not add any external imports. The only allowed imports are
from the Go standard library.

**Single responsibility:** Each file owns one concern -- emoji, normalization, or
the sanitization pipeline. Keep it that way.

**Compiled regexes only:** All `regexp.MustCompile` calls must be at package-level
`var` declarations, never inside functions. This keeps execution fast and predictable.

**No returned errors (pure text APIs):** The core text-processing functions
(`Demojify`, `ContainsEmoji`, `Normalize`, `Sanitize`) accept a string and return
a string or bool. They never return errors; panics are only permitted in
`regexp.MustCompile` at package-init time.

**Scanner functions are the exception:** `ScanDir` and `ScanFile` perform file I/O
and return `error` for filesystem failures (unreadable files, bad root path, etc.).
New scanner functions must also return errors for I/O failures. Do not add error
returns to the pure text-processing functions.

## Code Patterns

### Adding a regex-based rule

```go
// Package-level var -- compiled once at init time.
var myRuleRE = regexp.MustCompile(`(?im)^pattern[ \t]*\n?`)

func applyMyRule(text string) string {
    return myRuleRE.ReplaceAllString(text, "")
}
```

### Extending Options

```go
type Options struct {
    // ...existing fields...

    // RemoveNewFeature describes what this option does.
    RemoveNewFeature bool
}

func DefaultOptions() Options {
    return Options{
        // ...existing fields...
        RemoveNewFeature: true, // enable by default if generally useful
    }
}
```

### Pipeline order in Sanitize

Steps must follow this order:
1. `RemoveEmojis` -- structural characters removed first
2. `NormalizeWhitespace` -- final pass; collapses gaps left by emoji removal

## Unicode Emoji Ranges

Defined in `demojify.go` as `emojiRE`. When adding a new block:
1. Verify the range against [Unicode Charts](https://unicode.org/charts/)
2. Add the hex range in the same format: `\x{NNNN}-\x{NNNN}` or single `\x{NNNN}`
3. Keep the range list sorted by codepoint
4. Update the coverage table in `README.md`

## Code Style

```bash
# Format
gofmt -s -w .

# Vet
go vet ./...

# Lint (requires golangci-lint)
golangci-lint run ./...
```

**Rules:**
- `gofmt -s` formatting
- Exported symbols must have godoc comments
- Package-level vars before functions
- No `init()` functions; use `var` for side-effect-free initialization

## Performance Guidelines

1. **Regex at init time** -- compile all regular expressions as package-level `var`
2. **No allocations in hot paths** -- prefer `regexp.ReplaceAllString` (single pass)
3. **Benchmark before optimizing** -- use `go test -bench=. -benchmem`

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [regexp package docs](https://pkg.go.dev/regexp)
- [Unicode emoji charts](https://unicode.org/charts/)
