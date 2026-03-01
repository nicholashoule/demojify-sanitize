# Design Rationale

This document explains the key design decisions behind `demojify-sanitize`.
It is intended for contributors, enterprise evaluators, and developers of web
applications or APIs who use AI agents and need to understand how the module
audits, detects, and fixes content issues before they reach production.

## Zero-dependency policy

The library imports only the Go standard library (`os`, `path/filepath`,
`regexp`, `strings`, `unicode`, `unicode/utf8`).

**Why:** Every dependency in a shared library becomes a transitive dependency
for every project that imports it. In enterprise environments this creates
supply-chain audit burden, license-compatibility concerns, and upstream breakage
risk. Because all required operations (regex matching, string replacement,
trimming) are available in the stdlib, there is no benefit to accepting those
costs.

The policy is enforced by the absence of any `require` block in `go.mod`.

## Compiled package-level regexes

All `regexp.MustCompile` calls are `var` declarations at package scope, never
inside functions.

**Why:** Compiling a regex is relatively expensive. Compiling inside a function
means paying that cost on every call. Package-level `var` declarations run once
at program init — effectively free after that. `MustCompile` (rather than the
error-returning `Compile`) is safe here because the patterns are literals known
at compile time; a panic at init is preferable to silently carrying a nil regex
into production.

## No returned errors

The primary functions have the signature `func(string) string` or
`func(string) bool`; `Sanitize` additionally accepts an `Options` value.
None return `error`.

**Why:** The operations performed — regex replacement, string trimming — cannot
fail on valid UTF-8 input, and Go strings are always valid sequences of bytes
(even if not valid UTF-8, the regex engine handles the worst case gracefully).
Forcing callers to check an error that can never occur adds noise with no
benefit. If a regex were malformed, `MustCompile` panics at init, catching the
defect at development time rather than silently degrading at runtime.

## Pipeline order in `Sanitize`

Steps run in this fixed order:
1. Emoji removal (`Demojify`)
2. Whitespace normalization (`Normalize`)

**Why:** Emojis are removed first. An emoji adjacent to text leaves behind a
space; normalization cleans those up as a final pass rather than requiring each
step to tidy up after itself.

## External test package (`package demojify_test`)

All test files use `package demojify_test`, not `package demojify`.

**Why:** This enforces testing through the public API only, the same surface
that callers use. It catches cases where a function works internally but the
exported contract is wrong. It also means the test files serve as living
documentation of correct usage.

## Intentional Unicode exclusions

`©` (U+00A9), `®` (U+00AE), `™` (U+2122), mathematical arrows
(U+2190–U+2193), and all non-emoji scripts (CJK, Arabic, Latin extended, etc.)
are deliberately **not** removed.

**Why:** These symbols carry semantic meaning in technical and legal text.
Removing `®` or `©` from a product description or license notice would corrupt
the document. The library's contract is to remove *decorative* emoji, not all
non-ASCII characters. The Unicode emoji specification is the authoritative
source for which codepoints are emoji; ranges outside that specification are
left alone.

## File scanner and error handling

`ScanDir` and `ScanFile` are the only public functions that return an `error`.
The text-processing functions (`Demojify`, `Normalize`, `Sanitize`) cannot fail
on string input, so they omit errors entirely (see "No returned errors" above).

The scanner performs file I/O -- reading files, walking directory trees -- which
can fail for reasons outside the library's control (permissions, missing paths,
filesystem errors). Returning an error from these functions is the idiomatic Go
approach and does not weaken the library's error-handling contract.

`ScanConfig` provides three exemption axes -- directories (`SkipDirs`), files
(`ExemptFiles`), and suffixes (`ExemptSuffixes`) -- plus an extension filter
(`Extensions`). `DefaultScanConfig` returns safe defaults for a typical Go
module repo. The scanner reuses the same `Options` struct and `Sanitize` pipeline
that callers already know, keeping the API surface consistent.

## Scope boundaries

The library intentionally does not:
- Parse or transform Markdown syntax
- Detect or remove profanity or sensitive content
- Perform language detection
- Provide streaming or `io.Reader`/`io.Writer` interfaces

**Why:** Each of these would require either external dependencies or significant
scope expansion that would dilute the library's focused purpose. Projects
needing those capabilities should compose this library with purpose-built tools.
