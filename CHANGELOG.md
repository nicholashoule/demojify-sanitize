# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
