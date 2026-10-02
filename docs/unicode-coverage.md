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

### Script-aware removal (`Options.ScriptAware`, v1.1.0)

Two of the matched codepoints also belong to ordinary written language:

- **U+200D Zero Width Joiner.** Devanagari, Malayalam, Sinhala and other
  Indic scripts use it to choose a half form, a chillu or a conjunct
  (ka + virama + ZWJ + ssa is a different rendering from ka + virama + ssa).
- **U+FE00-U+FE0D standardized variation selectors.** CJK compatibility
  ideographs, Mongolian and mathematical symbols use them to pick a glyph.

By default both are removed wherever they appear, and `ContainsEmoji`
reports them, so Indic text with a joiner reads as containing emoji. With
`Options.ScriptAware` (and `ContainsEmojiWith`):

| Codepoint | Removed when | Kept when |
|---|---|---|
| U+200D ZWJ | an emoji is on either side (skipping variation selectors, so `heart + U+FE0F + ZWJ + fire` loses all of it) | it joins anything else |
| U+FE00-U+FE0D | it follows an emoji | it follows anything else |
| U+FE0E, U+FE0F presentation selectors | always, as by default | never: they only choose text or emoji presentation, and removing them leaves the base in its text form (the digit of a keycap, the plain copyright sign) |
| every other matched codepoint | always, as by default | never |

ScriptAware never removes anything the default keeps: its output differs from
the default's only by joiners and variation selectors it kept, and a test
checks that over 20,000 random strings. It is opt-in, so the default behavior
of every function is unchanged. It applies to `Sanitize` and the functions
built on it (including the scanner's `Options`), and to `ContainsEmojiWith`;
`Demojify`, `ContainsEmoji`, `CountEmoji` and the `Replace` family keep the
default rules.

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
The exception is the joiner and the variation selectors those scripts share
with emoji sequences: removed everywhere by default, kept in ordinary text
with `Options.ScriptAware` (see above).

### Currency and letterlike symbols

Symbols like U+20AC (Euro sign) and U+00B0 (degree sign) are not emoji and are
not removed.

## Substitution vs. stripping

`Demojify` and `Sanitize` always strip; neither consults a replacement map.
To preserve meaning, use `DefaultReplacements()` with `Replace`, `ReplaceFile`,
or `ScanConfig.Replacements`. Each substitutes mapped sequences and then strips
any residual unmapped emoji in the same pass, so no separate `Demojify` call is
needed.

The `-sub` flag in the CLI does exactly this: substitutes known emoji with text
tokens, then removes any residual unmapped codepoints.

`DefaultReplacements()` covers 280 codepoint sequences across 24 categories:

1. Warning and Alerts
2. Status Symbols
3. Information
4. CI/CD Workflow
5. Favorites and Highlights
6. Cloud and Deployment
7. Project and Issue Tracking
8. Community and Contributors
9. Status Indicators
10. Severity (colored circles)
11. Stop and Prohibition
12. Platform and Language Indicators
13. Colored Squares
14. Arrows
15. Media Controls
16. Heart Variants
17. Math Operators
18. Geometric Shapes
19. Checkboxes
20. Common Dingbats
21. Calendar and Date Indicators
22. Scissors / Removed
23. Deprecated
24. Flags

See [replacements.md](replacements.md) for the full substitution table.

## Checking coverage programmatically

```go
// Check whether a specific codepoint would be removed:
removed := demojify.Demojify(string(r)) == ""

// Check whether text contains any removable codepoints:
hasEmoji := demojify.ContainsEmoji(text)
```
