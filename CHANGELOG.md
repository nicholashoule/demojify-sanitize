# Changelog

Notable changes are documented following [Keep a Changelog] and
[Semantic Versioning].

## [Unreleased]

Intended release: v1.2.0 (minor: additive API plus fixes). Nothing an
existing caller relies on changes: every v1 default (a zero `Options`
passing text through, `DefaultOptions`, `DefaultScanConfig` and
`ContainsEmoji` without `ScriptAware`, `MaxFileBytes` zero meaning no
limit) and every exported signature stays as it was.

### Added

- `SafeOptions()`: the recommended configuration, emoji removal with
  `ScriptAware` and whitespace left alone, so it is safe for any text,
  including JSON string values and source code. Add `NormalizeWhitespace`
  or allow lists on the returned value as needed.
- `SanitizeSafe(text)`: `Sanitize` with `SafeOptions()`, the recommended
  entry point: no `Options` value to leave unset by mistake.
- `SafeScanConfig()`: `DefaultScanConfig()` with `SafeOptions()`, the
  recommended scanner preset.
- `DefaultMaxFileBytes` (1 MiB), the limit both presets set in
  `ScanConfig.MaxFileBytes`.
- `ErrFileChanged`, returned (wrapped) by `WriteFinding`, `FixDir`,
  `SanitizeFile` and `ReplaceFile` when a file changed after it was read.
- Native fuzz tests (`FuzzSanitizeSafe`, `FuzzAllowedEmojis`, with seed
  corpora), examples (`ExampleSanitizeSafe`, `ExampleSafeOptions`) and
  regression tests for every fix below.

### Changed

- Documentation recommends `SanitizeSafe`, `SafeOptions` and
  `SafeScanConfig`, and describes the v1 defaults as stable but easy to
  misuse (a forgotten `Options` field sanitizes nothing). The README's
  pre-commit hook scripts moved to [docs/git-hooks.md](docs/git-hooks.md),
  which already carried them. The README warns against running `Sanitize`
  over raw JSON bytes (use `SanitizeJSON`).
- Benchmarks use `b.Loop()`; the `modernize` linter is enabled.

### Fixed

- The `demojify` CLI scans with `SafeScanConfig()` instead of
  `DefaultScanConfig()`, so its emoji removal is script-aware. Before, an
  audit could report, and `-fix` would delete, a Zero Width Joiner that
  shapes script text: `-fix` turned the Devanagari half-form (U+0915 U+094D
  U+200D U+0937) into the conjunct (U+0915 U+094D U+0937), silently changing
  how the text renders. Emoji, and the
  joiners inside emoji sequences, are removed exactly as before; flags and
  exit codes are unchanged. The library's v1 defaults (`DefaultOptions`,
  `DefaultScanConfig`) are untouched.
- `WriteFinding` and `FixDir` (and `SanitizeFile`, `ReplaceFile`) no
  longer overwrite a file that changed after it was read: just before the
  rename the target is compared with the content that was cleaned, and a
  mismatch leaves the file alone and returns an error wrapping
  `ErrFileChanged`. A `Finding` held across an edit used to replace the edit
  with content cleaned from the stale copy.
- `ScanConfig.MaxFileBytes` bounds the read, not only the size check. The
  check uses the walk entry's own size, which for a symlink is the link's,
  so a symlink to a file over the limit was read whole; a file that grew
  after the check was too. Both are now skipped.

### Security

- `Options.AllowedEmojis` no longer swaps placeholders into the text. Input
  holding all 34 Unicode noncharacters and the fallback placeholder sequence
  had that text turned into an allowed emoji, and an allowed string matching
  inside a placeholder (an ASCII digit) corrupted the output. The allowed
  occurrences are now protected by marking their bytes in place, with the
  same longest-first, leftmost, non-overlapping choice, so ordinary input
  gives exactly the previous result (checked by `FuzzAllowedEmojis` against
  the previous implementation) and hostile input can no longer collide.

## [1.1.0] - 2026-10-02

Additive only: with the new option off, every function returns exactly what
v1.0.0 returned. A differential test against v1.0.0 found no difference over
200,000 random inputs through `Demojify`, `ContainsEmoji`, `CountEmoji`,
`BytesSaved`, `Normalize`, `Replace`, `FindAll`, `Sanitize`,
`SanitizeReport` and `SanitizeJSON` across the existing option sets.

### Added

- `Options.ScriptAware` (off by default) keeps two codepoints that ordinary
  written language shares with emoji sequences:
  - the Zero Width Joiner (U+200D), which Devanagari, Malayalam, Sinhala and
    other Indic scripts use for half forms, chillus and conjuncts;
  - the standardized variation selectors (U+FE00-U+FE0D), which CJK,
    Mongolian and mathematical text use to pick a glyph.

  Each is still removed inside an emoji sequence (a joiner with an emoji on
  either side, a selector after an emoji). The presentation selectors
  U+FE0E and U+FE0F, and every other emoji codepoint, are removed as before.
  An emoji kept by `AllowedEmojis` still counts as an emoji for its
  neighbors, so a joiner or selector beside it is removed.
  The option applies to `Sanitize` and everything built on it
  (`SanitizeReport`, `SanitizeReader`, `SanitizeJSON`, `SanitizeFile`, and
  the scanner through `ScanConfig.Options`). See
  [docs/unicode-coverage.md](docs/unicode-coverage.md).
- `ContainsEmojiWith(text, opts)`: whether `Sanitize` with opts would remove
  anything, honoring `ScriptAware`, `AllowedRanges` and `AllowedEmojis`. With
  a zero `Options` it answers exactly as `ContainsEmoji`. An input gate can
  now accept Indic text that `ContainsEmoji` reports as emoji.
- Examples: `ExampleSanitize_scriptAware`, `ExampleContainsEmojiWith`, and a
  ScriptAware section in `docs/examples/driver`.

### Changed

- The minimum Go version is 1.24 (was 1.21, which reached end of life in
  2024). CI tests Go 1.24, 1.25 and the current stable release; the macOS
  exclusion that only Go 1.21 needed is gone. Toolchains since Go 1.21
  download a newer toolchain automatically when a module asks for one.
- With `ScriptAware`, `Finding.HasEmoji` reports what `ContainsEmojiWith`
  reports, so a file whose only matches are joiners in ordinary text is not
  a finding. Without the option it is unchanged.
- Documentation: install and CI examples pin `@v1.1.0`. The coverage notes
  no longer claim Devanagari text is always untouched: by default its
  joiners are removed. Releases before 1.0.0 are grouped as legacy releases,
  here and in SECURITY.md.

## [1.0.0] - 2026-09-14

First stable release. The public Go API, the CLI flags and exit codes, and the
`-json` output schema are now covered by semantic versioning: no breaking
changes without a major version bump. The human-readable text output, the set
of Unicode ranges treated as emoji (which tracks new Unicode releases), and the
contents of `DefaultReplacements()` may still grow or change in minor releases.

### Changed

- Library behavior is identical to v0.10.1. The module path is unchanged (no
  `/v1` suffix), so upgrading requires no code changes.
- CLI: the clean-tree message is now `[PASS] no findings` (previously
  `[PASS] no emoji found`), since `-normalize` runs can report whitespace-only
  findings. Scripts should rely on the exit code or `-json`, not this text.
- The repository's own pre-commit hook runs the CLI from the working tree
  instead of a published release, so it always exercises the current code.
- Documentation: install and CI examples pin `@v1.0.0`; package, CLI, and
  design docs now describe file replacement as atomic on POSIX and best-effort
  on Windows, and clarify Unicode coverage and error behavior.

## Legacy releases (before 1.0.0)

Releases before 1.0.0 are legacy and unsupported ([SECURITY.md](SECURITY.md)).
They predate the v1 compatibility promise, so their API and behavior could
change between minor versions. Upgrade to 1.x: the 1.0.0 library behaves
exactly as 0.10.1 did, so the move needs no code change.

Notes for 0.10.0 and earlier are in this file's git history and in these tag
comparisons: [0.10.0], [0.9.0], [0.8.0], and [0.7.3 and earlier].

[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.10.1...v1.0.0
[0.10.1]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.10.0...v0.10.1
[0.10.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.3...v0.8.0
[0.7.3 and earlier]: https://github.com/nicholashoule/demojify-sanitize/releases/tag/v0.7.3
