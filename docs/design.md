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

## Substitution pipeline

The `Replace` family of functions (`Replace`, `ReplaceFile`, `ReplaceCount`,
`FindAllMapped`, `FindMatchesInFile`) addresses a common pattern: rather than
silently removing emoji, callers want to map them to readable text equivalents
(e.g., `[PASS]`, `WARNING`, `->`) so that context is preserved in plain-text
output.

**Why `Replace` delegates to `Demojify` after substitution:**
The replacement map is curated and finite. Inputs may contain emoji outside
the map -- especially supplementary block emoji (U+1F000–U+1FAFF) added in
recent Unicode versions. Rather than silently producing garbled output,
`Replace` passes the substituted text through `Demojify` as a residual cleanup
step. Callers get a clean string regardless of whether every codepoint was in
their map.

**Why longest-key matching is required:**
Many emoji appear in both a bare form (e.g., U+26A0 WARNING SIGN) and a
variation-selector form (U+26A0 U+FE0F). If the bare codepoint were matched
first, the variation selector U+FE0F would remain and be stripped by `Demojify`
as a residual, leaving a stray space or no-op character. Processing keys in
descending byte-length order (via `strings.NewReplacer` for `Replace`, and a
explicit linear scan for `FindAllMapped`) ensures multi-codepoint sequences are
always consumed atomically.

**Why `DefaultReplacements()` returns a copy:**
A shared global map is not safe for concurrent mutation. Returning a fresh
copy on every call lets each caller add, remove, or override entries without
affecting other goroutines or call sites. The copy cost is negligible (~137
entries) compared to the I/O in `ReplaceFile` or the regex in `Demojify`.

**Why `ReplaceFile` uses an atomic rename:**
Writing directly to the target file leaves a window where a crash or
interruption would produce a truncated file. Writing to a sibling temp file
and then calling `os.Rename` means the file is either fully updated or fully
unchanged.

On POSIX systems `rename(2)` is atomic and replaces the destination in a
single filesystem operation. On Windows, Go 1.21+ (the minimum version for
this module) implements `os.Rename` via `MoveFileEx` with
`MOVEFILE_REPLACE_EXISTING`, which replaces the destination file but is **not**
guaranteed to be atomic by the Windows kernel -- a crash during the move could
theoretically leave the destination absent. In practice this is safe for
single-file replace-in-place on the same volume (the temp file is always
created in the same directory). Cross-volume renames are not attempted.

## Streaming sanitization

`SanitizeReader(r io.Reader, w io.Writer, opts Options) error` applies the
same pipeline as `Sanitize` line by line against an `io.Reader`, writing
results to an `io.Writer`. It is designed for streaming scenarios -- LLM token
streams, MCP transport payloads, HTTP chunked responses -- where buffering the
complete input is undesirable. All options (emoji removal, whitespace
normalization, allowed ranges/emojis) are honoured per line.

The internal `bufio.Scanner` is configured with a 1 MiB per-line buffer
(`sanitizeReaderMaxTokenSize`). This accommodates minified JSON, base64-encoded
payloads, and long LLM output lines that would exceed the default 64 KiB scanner
limit. Lines longer than 1 MiB cause `bufio.ErrTooLong` to be returned.

## JSON sanitization

`SanitizeJSON(data []byte, opts Options) ([]byte, error)` sanitizes only the
string values within a JSON document, leaving keys, numbers, booleans, and null
untouched. It uses `json.Decoder` with `UseNumber` to preserve numeric precision,
and after decoding the first value it performs a second `Decode` to verify EOF is
reached. Inputs with trailing non-whitespace data (e.g., `{"a":1} trailing` or
two concatenated JSON objects) are rejected with an error, ensuring the caller
always receives a single, complete, structurally valid JSON document.

## Batch scan-and-fix

`FixDir(root string, cfg ScanConfig) (fixed, clean int, err error)` is the
write-side complement to `ScanDir`. It walks the directory tree at `root`,
applies the sanitization or replacement pipeline from `cfg`, and atomically
writes back every file whose content changed. It returns counts of fixed and
already-clean files. Path-traversal protection (via `filepath.EvalSymlinks`
and `isInsideDir`) ensures no write target can escape `root` through `..`
components or symlinks.

## Scope boundaries

The library intentionally does not:
- Parse or transform Markdown syntax
- Detect or remove profanity or sensitive content
- Perform language detection

**Why:** Each of these would require either external dependencies or significant
scope expansion that would dilute the library's focused purpose. Projects
needing those capabilities should compose this library with purpose-built tools.

## Per-file line limit configuration

`LimitConfig`, `DefaultLimitConfig()`, and `ResolveLimit()` provide a lightweight mechanism for
expressing per-file line limit policies that external governance tooling can
apply when deciding how many lines of a file to inspect.

**Important:** The core scanner APIs (`ScanDir`, `ScanDirContext`, `ScanFile`)
do not currently consume `LimitConfig` directly. Callers that need hard
per-file line limits should enforce those limits in their own orchestration
layer (for example, by truncating content or skipping files) based on
`LimitConfig` before or around calls into the scanner.

**Why a separate config type instead of a field on `ScanConfig`:**
`ScanConfig` controls what to scan (root, skip dirs, extensions) and how to
process it (options, replacements). Line limits are a governance concern --
they express policy about file size -- and belong in a dedicated type. This
keeps `ScanConfig` focused and lets callers compose their own governance
policies without coupling them to the scan pipeline's implementation details.

**Why the zero-value of `Default` falls back to `DefaultLineLimit`:**
A zero `Default` is indistinguishable from "not set by the caller" when the
struct is value-initialized. Treating zero as a sentinel avoids a separate
`bool` sentinel field and lets callers use `LimitConfig{}` to mean
"apply the standard 500-line default" without explicitly knowing the constant.

**Why `.claude/CLAUDE.md` has a built-in 50-line override:**
AI context files (CLAUDE.md, AGENTS.md, and similar) are designed to be
short, focused instruction sets. A 50-line override in `DefaultLimitConfig()`
expresses this constraint out of the box for the most common AI-agent workspace
layout, giving projects a governance nudge without requiring manual
configuration.