# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `DefaultReplacements()` expanded from ~97 to ~135 entries across eleven
  categories. New entries:
  - **Information symbol** -- U+2139 -> `[INFO]`
  - **Severity circles** -- U+1F534 -> `[ERROR]`, U+1F7E0 -> `[WARNING]`,
    U+1F7E1 -> `[CAUTION]`, U+1F7E2 -> `[OK]`, U+1F535 -> `[INFO]`,
    U+26AB/U+26AA (medium circles) -> `[INACTIVE]`
  - **Stop / prohibition** -- U+1F6D1 -> `[STOP]`, U+26D4 -> `[NO ENTRY]`,
    U+1F6AB -> `[PROHIBITED]`
  - **CI/CD workflow** -- U+1F680 -> `[DEPLOY]`, U+1F4E6 -> `[PACKAGE]`,
    U+1F389 -> `[SUCCESS]`, U+2728 -> `[NEW]`, U+1F3C1 -> `[DONE]`,
    U+1F527 -> `[FIX]`, U+1F6E0 -> `[TOOLS]`, U+267B -> `[RECYCLE]`,
    U+1F4BE -> `[SAVE]`, U+1F525 -> `[HOT]`, U+1F4AF -> `[100]`
  - **Math operators** -- U+2716 -> `x`, U+2795 -> `+`, U+2796 -> `-`,
    U+2797 -> `/`, U+267E -> `[INFINITY]`
  - **Expanded FE0F variants** -- added variation-selector-suffixed forms for
    U+2705, U+274C, U+2757, U+2755, U+23F3, U+231B, U+23F1
- `emojiRE` in `demojify.go` extended with two new ranges:
  - **U+2139** -- Information Source, previously undetected by `Demojify` /
    `ContainsEmoji`
  - **U+E0020–U+E007F** -- Tags block; covers subdivision flag sequences
    (England, Scotland, Wales) that slipped through before
- `Options.AllowedEmojis []string` -- preserves exact emoji strings during
  removal (complements `AllowedRanges` which works at the Unicode block level)

### Changed (previously listed)

- `Normalize` rewritten to preserve leading indentation on each line. Only
  consecutive spaces/tabs after the first non-whitespace character are collapsed.
  This makes `-normalize` safe for Markdown nested lists, indented code blocks,
  and aligned source comments.
- `ScanConfig` struct -- configures directory/file exemptions (`SkipDirs`,
  `ExemptFiles`, `ExemptSuffixes`), extension filters, and sanitization
  `Options` for file-level scanning.
- `DefaultScanConfig()` -- returns a config suitable for Go module and AI-agent
  repos (skips `.git/`, `vendor/`, `node_modules/`; exempts `*_test.go`; scans
  all file types by default). Set `Extensions` to restrict to specific types.
- `ScanDir(ScanConfig) ([]Finding, error)` -- walks a directory tree and
  returns a `Finding` for every file whose content would change after
  sanitization.
- `ScanFile(path, Options) (*Finding, error)` -- checks a single file and
  returns a `Finding` or nil if already clean.
- `Finding` struct with `Path`, `HasEmoji`, `Original`, and `Cleaned` fields.
- `Options.AllowedRanges []*unicode.RangeTable` -- preserves specific emoji
  codepoints during removal while stripping all others (backward-compatible;
  `nil` default removes everything as before).
- `repo_test.go` -- dogfooding tests that validate the entire repository using
  the module's own API (`ContainsEmoji`, `Sanitize`). Five tests:
  `TestRepoProductionSourceFilesEmojiClean`, `TestRepoAllDocsEmojiClean`,
  `TestRepoProductionFilesIdempotent`, `TestRepoTestFilesContainEmoji`,
  `TestRepoAgentOutputRemediation`.
- `make fmt-check` target and CI `Format check` step enforcing `gofmt -s`.
- Apache License 2.0 (`LICENSE` file).
- Windows note in `Makefile` and `CONTRIBUTING.md`: race detector requires
  CGO and gcc.

[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.1.0...HEAD
