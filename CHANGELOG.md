# Changelog

Notable changes are documented following [Keep a Changelog] and
[Semantic Versioning].

## [Unreleased]

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
- Documentation: install and CI examples pin `@v1.0.0`; package, CLI, and
  design docs now describe file replacement as atomic on POSIX and best-effort
  on Windows, and clarify Unicode coverage and error behavior.

## [0.10.1] - 2026-09-11

### Fixed

- `SanitizeReader` accepts lines of exactly 1 MiB; longer lines still return
  `bufio.ErrTooLong`.
- Match context for CRLF files no longer includes the trailing `\r`.
- Restored a valid POSIX pre-commit invocation and pinned the CI lint job's
  Go version to match its linter build.

### Changed

- Replacement scanning buckets keys by first byte and advances unmatched text
  by rune, reducing prefix checks.
- The repository pre-commit hook scans more text file types and skips
  `build/`.

## Older Releases

Notes for earlier versions are in this file's git history and in these tag
comparisons: [0.10.0], [0.9.0], [0.8.0], and [0.7.3 and earlier].

[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.10.1...v1.0.0
[0.10.1]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.10.0...v0.10.1
[0.10.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.3...v0.8.0
[0.7.3 and earlier]: https://github.com/nicholashoule/demojify-sanitize/releases/tag/v0.7.3
