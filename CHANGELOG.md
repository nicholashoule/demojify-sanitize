# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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

## [0.1.0] - 2026-02-28

### Added

- `Demojify(text string) string` -- removes emoji and Unicode pictographic
  characters using a compiled regex covering Unicode 15 ranges.
- `ContainsEmoji(text string) bool` -- reports whether text contains an emoji.
- `Normalize(text string) string` -- collapses redundant whitespace and blank lines.
- `Sanitize(text string, opts Options) string` -- configurable pipeline combining
  all three operations in order: emoji removal, AI-clutter removal, normalization.
- `Options` struct with `RemoveEmojis`, `RemoveAIClutter`, `NormalizeWhitespace` fields.
- `DefaultOptions()` returns all fields set to `true`.
- AI-clutter removal for 13 common preamble phrases (case-insensitive, line-anchored).
- Runnable `Example*` functions verified by `go test`.
- Zero external dependencies; requires Go 1.21+.
- CI workflow (GitHub Actions) with race detector and coverage reporting.

[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/nicholashoule/demojify-sanitize/releases/tag/v0.1.0
