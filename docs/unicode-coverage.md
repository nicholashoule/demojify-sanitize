# Unicode Coverage

This document describes which Unicode codepoints `Demojify` (and by extension
`Sanitize`) removes, which are intentionally excluded, and why.

## What is removed

`Demojify` removes every codepoint matched by the internal `emojiRE` regex.
The regex is a character class covering the Unicode emoji/pictographic
assignments through Unicode 17.

| Range | Description |
|-------|-------------|
| U+2139 | Information Source |
| U+231A–U+231B | Watch, Hourglass Done |
| U+23CF | Eject Symbol |
| U+23E9–U+23F3 | Fast-forward through Hourglass (media controls) |
| U+23F8–U+23FA | Pause, Stop, Record buttons |
| U+24C2 | Circled M |
| U+25AA–U+25AB | Black/White Small Square |
| U+25B6 | Black Right-Pointing Triangle (play button) |
| U+25C0 | Black Left-Pointing Triangle (reverse button) |
| U+25FB–U+25FE | Medium squares |
| U+2600–U+27BF | Miscellaneous Symbols and Dingbats (sun, moon, stars, arrows, checkmarks, ...) |
| U+2934–U+2935 | Curved arrows |
| U+2B05–U+2B07 | Directional arrows (left, up, down) |
| U+2B1B–U+2B1C | Large Black/White Square |
| U+2B50 | White Medium Star |
| U+2B55 | Heavy Large Circle |
| U+3030 | Wavy Dash |
| U+303D | Part Alternation Mark |
| U+3297 | Circled Ideograph Congratulation |
| U+3299 | Circled Ideograph Secret |
| U+1F000–U+1FAFF | All supplementary emoji blocks: Mahjong tiles, dominoes, playing cards, enclosed alphanumerics, transport/map symbols, miscellaneous symbols, emoticons, skin-tone modifiers, geometric shapes, supplemental arrows, supplemental symbols and pictographs, chess symbols, symbols and pictographs extended-A (includes all Emoji 17.0 additions) |
| U+200D | Zero Width Joiner (used in multi-part emoji sequences; stripped individually) |
| U+20E3 | Combining Enclosing Keycap |
| U+E0020–U+E007F | Tags block (tag characters used in subdivision flag sequences: England, Scotland, Wales) |
| U+FE00–U+FE0F | Variation Selectors 1–16 (the FE0F emoji presentation selector) |

### How multi-codepoint sequences are handled

The regex matches and removes individual codepoints. It does not understand
multi-codepoint sequences as atomic units. This means:

- A ZWJ sequence like [woman] + U+200D + [laptop] is stripped codepoint by
  codepoint: the woman emoji, the ZWJ, and the laptop emoji are each removed.
- A subdivision flag like [U+1F3F4 U+E0067 U+E0062 U+E0065 U+E006E U+E0067
  U+E007F] (England) is also stripped completely because both U+1F3F4
  (Waving Black Flag) and the U+E0020–U+E007F tag range are covered.
- Skin-tone modifier codepoints (U+1F3FB–U+1F3FF) fall within U+1F000–U+1FAFF
  and are stripped.

The only artifact this can leave is whitespace between the words that
surrounded the emoji. Use `Normalize` (or `-normalize` in the CLI) to clean
those up.

## What is intentionally NOT removed

The following codepoints are explicitly out of scope.

### Legal and trademark symbols

| Codepoint | Symbol | Reason |
|-----------|--------|--------|
| U+00A9 | (c) | Copyright notice -- legally significant in documents, licenses, and source code headers |
| U+00AE | (R) | Registered trademark -- legally significant |
| U+2122 | TM | Trademark symbol -- legally significant |

Removing these from a legal notice, license file, or product documentation
would corrupt the document's meaning.

### Mathematical and technical arrows

| Range | Description | Reason |
|-------|-------------|--------|
| U+2190–U+2193 | Basic directional arrows (left, up, right, down) | Widely used in mathematical notation, type theory, data-flow diagrams, and technical documentation |
| U+21D0–U+21D3 | Double arrows | Logical implication in math and type systems |

These are not emoji. They appear in Unicode's "Arrows" block (U+2190–U+21FF),
which predates emoji and is used extensively in academic and technical writing.

Note that `DefaultReplacements()` maps U+2192 (`->`) and related arrows so
they can be substituted in documentation pipelines via `Replace`/`-sub`. This
is opt-in: `Demojify` alone does not touch them.

### All non-emoji Unicode scripts and blocks

CJK (Chinese, Japanese, Korean), Arabic, Hebrew, Cyrillic, Latin Extended,
Greek, Devanagari, and all other writing-system codepoints are untouched.
The library targets decorative pictographic content, not written language.

### Currency and letterlike symbols

Symbols like U+20AC (Euro sign) and U+00B0 (degree sign) are not emoji and are
not removed.

## Substitution vs. stripping

`Demojify` always strips. To preserve meaning, use `DefaultReplacements()` with
`Replace` or `ReplaceFile` before calling `Demojify` -- or use `Sanitize` with
`DefaultOptions()`, which runs the replacement step first.

The `-sub` flag in the CLI does exactly this: substitutes known emoji with text
equivalents, then `Demojify` removes any residual unmapped codepoints.

`DefaultReplacements()` covers approximately 230 codepoint sequences across
eighteen categories:

1. Warning and Alerts
2. Status Symbols
3. Information
4. Severity (colored circles)
5. Stop and Prohibition
6. Favorites and Highlights
7. Cloud and Deployment
8. CI/CD Workflow
9. Status Indicators
10. Arrows
11. Math Operators
12. Geometric Shapes
13. Checkboxes
14. Common Dingbats
15. Heart Variants
16. Project and Issue Tracking
17. Colored Squares
18. Media Controls
19. Community and Contributors
20. Platform and Language Indicators

See [replacements.md](replacements.md) for the full substitution table.

## Checking coverage programmatically

```go
// Check whether a specific codepoint would be removed:
removed := demojify.Demojify(string(r)) == ""

// Check whether text contains any removable codepoints:
hasEmoji := demojify.ContainsEmoji(text)
```
