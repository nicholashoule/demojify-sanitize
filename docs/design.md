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
(`Extensions`), an optional `Replacements` map (uses `Replace` instead of
`Sanitize` per file when set), and a `CollectMatches` flag (populates
`Finding.Matches` with per-occurrence detail). `DefaultScanConfig` returns
safe defaults for a typical Go module repo. The scanner reuses the same
`Options` struct and `Sanitize` pipeline that callers already know, keeping
the API surface consistent.

**Why CRLF line endings are preserved:**
The scanner internally normalizes `\r\n` to `\n` before running inline-space
cleanup, because the trailing-space regex and space-collapse logic must treat
all platforms consistently. After cleanup, the original line-ending convention
is restored: if the source file used `\r\n`, the output uses `\r\n`. This
ensures that removing emoji from a Windows-native file does not silently convert
it to LF line endings, which would produce noisy git diffs unrelated to the
actual emoji changes. The restoration only applies when `NormalizeWhitespace`
is false; callers who request full whitespace normalization accept LF output.
A file counts as CRLF only when every line break is `\r\n` -- a stray bare
`\r` or `\n` marks the file as mixed, and mixed files are left with LF so the
scanner never spreads `\r\n` beyond the file's own convention.

**Why the non-normalize cleanup is scoped to changed lines:**
When `NormalizeWhitespace` is false, the scanner still tidies the whitespace
artifacts that emoji removal leaves behind (two spaces where an emoji sat,
trailing spaces at line end). That cleanup compares the file line by line
against the original and touches only the lines emoji removal actually
changed. Running the collapse over the whole file -- the original behavior --
meant one emoji anywhere destroyed gofmt tab alignment and column-aligned
comments on every other line, so a "fix" pass left Go files unformatted.
Untouched lines are now preserved byte for byte.

## Substitution pipeline

The `Replace` family of functions (`Replace`, `ReplaceFile`, `ReplaceCount`,
`FindAllMapped`, `FindMatchesInFile`) addresses a common pattern: rather than
silently removing emoji, callers want to map them to readable text equivalents
(e.g., `[PASS]`, `WARNING`, `->`) so that context is preserved in plain-text
output.

**Why `Replace` is a single position-aware scan:**
`Replace` walks the input left to right once. At each position it tries the
replacement keys longest-first; if none match, it checks for an unmapped emoji
codepoint (stripped) and otherwise copies the input byte through. This design
has three consequences that a find-and-replace pipeline cannot deliver:

1. **Unmapped emoji are still removed.** The replacement map is curated and
   finite; inputs may contain emoji outside it -- especially supplementary
   block emoji (U+1F000–U+1FAFF) added in recent Unicode versions. The scan
   strips them in the same pass, so callers get a clean string regardless of
   whether every codepoint was in their map.
2. **Replacement values are emitted verbatim.** A value is never re-scanned
   for emoji, so identity mappings (key == value) preserve the mapped
   codepoints exactly -- the documented way to keep specific codepoints during
   replacement-based scans.
3. **Only substituted tokens collapse.** Because the scan knows which output
   spans came from substitution, runs of adjacent repeated emoji collapse to a
   single token without ever touching literal input text that happens to equal
   a replacement token. The earlier whole-output collapse corrupted emoji-free
   documents that legitimately contained repeated token text such as
   `[WARNING] [WARNING]` in example CLI output.

**Why longest-key matching is required:**
Many emoji appear in both a bare form (e.g., U+26A0 WARNING SIGN) and a
variation-selector form (U+26A0 U+FE0F). If the bare codepoint were matched
first, the variation selector U+FE0F would remain and be stripped as a
residual, leaving a stray space or no-op character. Trying keys in descending
byte-length order at each position (the same greedy walk used by
`FindAllMapped`) ensures multi-codepoint sequences are always consumed
atomically.

**Why `DefaultReplacements()` returns a copy:**
A shared global map is not safe for concurrent mutation. Returning a fresh
copy on every call lets each caller add, remove, or override entries without
affecting other goroutines or call sites. The copy cost is negligible (~280
entries) compared to the I/O in `ReplaceFile` or the regex in `Demojify`.

**Why run collapsing skips tokens shorter than 4 characters:**
Several emoji in `DefaultReplacements()` map to short ASCII sequences: `/`
(U+2797 heavy division sign), `-` (U+2796 heavy minus), `*` (U+2022 bullet),
`o` (U+25CB white circle), `->` (U+2192 rightwards arrow). Two of those can
land adjacent in perfectly normal output (an arrow next to a bullet), and
deduplicating short sequences is far more likely to corrupt meaning than to
remove clutter. Only tokens of 4 or more bytes (e.g., `[FAIL]`, `[WARNING]`,
`WARNING`, `[DEPLOY]`) are label-like quantities where an adjacent repeat is
redundant, and even those collapse only when the run's last member was
produced by substitution -- literal input text is never removed.

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
and after decoding the first value it performs a second `Decode` into a
`json.RawMessage` to verify EOF is reached. Any second well-formed JSON value
(object, array, number, string, bool, null) returns `ErrMultipleJSONValues`;
trailing garbage that is not valid JSON (e.g., `{"a":1} trailing`) returns the
decoder's syntax error. Either way the caller only ever receives a single,
complete, structurally valid JSON document.

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

