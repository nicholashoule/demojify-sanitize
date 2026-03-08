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
| [demojify.go](../../demojify.go) | `Demojify`, `ContainsEmoji`, `CountEmoji`, `BytesSaved`, `TechnicalSymbolRanges` -- emoji detection/removal |
| [normalize.go](../../normalize.go) | `Normalize` -- whitespace normalization |
| [sanitize.go](../../sanitize.go) | `Sanitize`, `SanitizeFile`, `SanitizeReport`, `SanitizeReader`, `SanitizeJSON`, `Options`, `DefaultOptions`, `SanitizeResult` -- pipeline |
| [scan.go](../../scan.go) | `ScanConfig`, `DefaultScanConfig`, `ScanDir`, `ScanDirContext`, `ScanFile`, `FindMatchesInFile`, `Finding`, `Match` -- scanner |
| [replace.go](../../replace.go) | `Replace`, `ReplaceFile`, `ReplaceCount`, `FindAll`, `FindAllMapped` -- substitution |
| [replacements.go](../../replacements.go) | `DefaultReplacements` -- built-in emoji-to-text map |
| [write.go](../../write.go) | `WriteFinding` -- atomic write-back for scan results |
| [fix.go](../../fix.go) | `FixDir`, `isInsideDir` -- batch scan-and-fix |
| [helpers.go](../../helpers.go) | `isBinary`, `sortByLenDesc`, `sortedKeys`, `statAndWrite`, `collapseInlineSpaces` -- shared internals |
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

## Module Guidelines from pkg.go.dev

Source: [pkg.go.dev/about#adding-a-package](https://pkg.go.dev/about#adding-a-package)

pkg.go.dev auto-indexes modules via `proxy.golang.org`. Once a tagged version is
pushed to GitHub, the package appears within minutes -- no manual submission needed.

### Best-Practice Checklist

| Criterion | Status | Detail |
|-----------|--------|--------|
| `go.mod` present | PASS | `module github.com/nicholashoule/demojify-sanitize`, `go 1.21` |
| Redistributable license | PASS | Apache 2.0 (`LICENSE`) |
| `LICENSE` starts on line 1 | CHECK | pkg.go.dev `licensecheck` reads from byte 0 -- any leading blank line prevents detection; verify with `Get-Content LICENSE \| Select-Object -First 1` |
| Tagged version | PASS | `v0.1.0`, `v0.2.0` |
| Stable version (`v1+`) | **GAP** | Latest tag is `v0.2.0`; pkg.go.dev treats `v0.x` as experimental |
| Good package doc | PASS | `doc.go` opens with a one-sentence summary; all exported symbols have godoc |

### The Stable Version Gap

pkg.go.dev states:

> *"Projects at v0 are assumed to be experimental. When a project reaches a stable
> version -- major version v1 or higher -- breaking changes must be done in a new
> major version."*

Until `v1.0.0` is tagged:

- pkg.go.dev displays a stability warning on the package page.
- `go get` resolves pre-release semantics; consumers cannot rely on SemVer compatibility.
- The pkg.go.dev scorecard will not show the green "Stable version" check.

When the public API is locked, tag and push:

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Documentation Rules (pkg.go.dev)

- The **first sentence** of the package comment is indexed for search -- keep it clear and accurate.
- Every exported symbol must have a godoc comment; unexported symbols may be omitted.
- `example_test.go` functions appear as runnable examples on the pkg.go.dev page.

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [regexp package docs](https://pkg.go.dev/regexp)
- [Unicode emoji charts](https://unicode.org/charts/)
- [pkg.go.dev guide](https://pkg.go.dev/about#adding-a-package)
