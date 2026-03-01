# Design Rationale

This document explains the key design decisions behind `demojify-sanitize`.
It is intended for contributors and enterprise evaluators who need to understand
the *why*, not just the *what*.

## Zero-dependency policy

The library imports only the Go standard library (`regexp`, `strings`).

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

All public functions have the signature `func(string) string` or
`func(string) bool`. None return `error`.

**Why:** The operations performed — regex replacement, string trimming — cannot
fail on valid UTF-8 input, and Go strings are always valid sequences of bytes
(even if not valid UTF-8, the regex engine handles the worst case gracefully).
Forcing callers to check an error that can never occur adds noise with no
benefit. If a regex were malformed, `MustCompile` panics at init, catching the
defect at development time rather than silently degrading at runtime.

## Pipeline order in `Sanitize`

Steps run in this fixed order:
1. Emoji removal (`Demojify`)
2. AI-clutter removal
3. Whitespace normalization (`Normalize`)

**Why:**
- Emojis removed first ensures that an emoji adjacent to a clutter phrase (e.g.
  `"Certainly! 🎉"`) does not prevent the phrase regex from matching. The phrase
  patterns anchor on line content, not on character classes.
- AI-clutter removal before normalization means the phrase removal may leave
  blank lines or trailing spaces behind; normalization cleans those up as a
  final pass rather than requiring each earlier step to tidy up after itself.

## False-positive prevention in AI clutter patterns

Short, common English words (`Sure`, `Great`, `Noted`, etc.) require trailing
punctuation (`[!,.]`) before they are removed.

**Why:** Without the punctuation requirement, `"Sure enough, the build passed"`
or `"Great work by the team"` would be silently truncated, corrupting legitimate
content. The punctuation signals that the word is being used as a standalone
filler phrase rather than as part of a sentence. Longer, structurally
distinctive phrases (`"I'd be happy to help"`) do not need this guard because
they are unlikely to appear as mid-sentence fragments.

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

## Scope boundaries

The library intentionally does not:
- Parse or transform Markdown syntax
- Detect or remove profanity or sensitive content
- Perform language detection
- Provide streaming or `io.Reader`/`io.Writer` interfaces

**Why:** Each of these would require either external dependencies or significant
scope expansion that would dilute the library's focused purpose. Projects
needing those capabilities should compose this library with purpose-built tools.
