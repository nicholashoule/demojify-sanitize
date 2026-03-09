# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2026-03-09

### Added

- `scripts/hooks/pre-commit.go`: add `checkTest` -- runs `go test ./...` as the
  final gate after `checkFmt`, `checkVet`, and `checkLint`; a failing test suite
  now blocks the commit with `[FAIL] go test (run: make test)`
- `README.md`: fill in `## Features` section with 11 bullet points covering the
  full public API surface; fix truncated `go get` command and wrong CLI install
  path (`repogov/cmd/demojify-sanitize` -> `demojify-sanitize/cmd/demojify`)

### Changed

- `scripts/hooks/pre-commit`, `README.md`, `docs/git-hooks.md`: bump repogov
  from `v0.2.0` to `v0.3.0`

### Removed

- `repo_test.go` six license-hygiene tests deleted (no longer needed now that
  the Apache-2.0 `LICENSE` file is stable and pkg.go.dev detects it correctly):
  `TestRepoLicenseStartsOnLineOne`, `TestRepoLicenseApache20Canonical`,
  `TestRepoLicenseApache20SectionOrder`, `TestRepoLicenseFilename`,
  `TestRepoLicenseNotEmpty`, `TestRepoLicenseApache20ExactPhrases`

## [0.5.0] - 2026-03-08

### Added

- `docs/git-hooks.md` cross-platform lightweight pre-commit examples for
  macOS/Linux (`sh`), Windows PowerShell (`pre-commit.ps1` + `sh` shim), and
  Windows Git for Windows (`sh.exe` -- same hook, no shim needed)
- `docs/examples/driver/main.go` step 18: `SanitizeFile` -- atomic
  single-file sanitization with idempotency verification; `ScanDirContext`
  renumbered to step 19 (no behavioral change)

### Changed

- `scripts/hooks/pre-commit.go`: add `checkLint` -- runs `golangci-lint run ./...`
  using the project's `.golangci.yml`; skips gracefully with a `WARNING:` when
  `golangci-lint` is not installed so the hook stays portable; wired into
  `main()` after `checkVet`
- `scripts/hooks/pre-commit`: integrate
  [repogov](https://github.com/nicholashoule/repogov) governance checks (line
  limits + layout) as a first-class gate; repogov runs via `go run` from the
  sibling `../repogov` directory and skips gracefully when absent; all three
  exit codes (`repogov_exit`, `demojify_exit`, `precommit_exit`) ORed into the
  final status; removed stale commented-out block from v0.4.0
- `scripts/hooks/pre-commit.go`: remove `checkRepogov` (moved to shell script);
  keep `checkFmt`, `checkVet`, and `checkLint` in the Go hook; drop unused `path/filepath` import
- `README.md` pre-commit hook section: restructured into **Option A**
  (pre-built binary, CI-friendly) and **Option B** (`go run` with repogov
  governance, recommended for in-repo hooks); updated `docs/git-hooks.md`
  reference to mention cross-platform examples
- `docs/git-hooks.md` CLI hook intro: note repogov pairing and graceful-skip
  behaviour so consumers understand the two tools are complementary
- `example_test.go` `ExampleSanitize_markdownFiles`: clarified doc comment --
  distinguishes per-file loop pattern from whole-directory `FixDir`; points
  consumers to `ExampleFixDir` for lower-friction one-call use
- `docs/replacements.md`: condensed intro paragraph and removed redundant
  section description to stay within the 300-line repogov limit (305 -> 298)

## [0.4.0] - 2026-03-08

### Added

- `CountEmoji(text string) int` -- count emoji codepoint occurrences; concurrency-safe
- `BytesSaved(text string) int` -- report the net byte reduction from removing emoji from the input
- `TechnicalSymbolRanges() []*unicode.RangeTable` -- pre-built range table for check marks,
  warning signs, gears, card suits, stars, and music notation; pass to `Options.AllowedRanges`
- `SanitizeResult` struct -- `Cleaned`, `EmojiRemoved`, and `BytesSaved` for observability pipelines
- `SanitizeReport(text string, opts Options) SanitizeResult` -- `Sanitize` pipeline with metrics
- `SanitizeReader(r io.Reader, w io.Writer, opts Options) error` -- streaming line-by-line sanitization
- `SanitizeJSON(data []byte, opts Options) ([]byte, error)` -- sanitize JSON string values only;
  preserves keys, numbers, booleans, and null; uses `json.Number` for numeric precision
- `ScanDirContext(ctx context.Context, cfg ScanConfig) ([]Finding, error)` -- context-aware
  directory scanner; stops on cancellation and returns collected findings
- `helpers.go` -- shared internal utilities (`isBinary`, `sortByLenDesc`, `sortedKeys`,
  `statAndWrite`, `collapseInlineSpaces`) extracted from production source files
- `docs/git-hooks.md` -- guide for dependency-free Git pre-commit hook integration
- README Git pre-commit hook pattern, `SanitizeReader` and `SanitizeJSON` examples
- `docs/design.md` streaming sanitization and batch scan-and-fix sections
- Concurrency tests for `Demojify`, `Sanitize`, `Replace`, and `ScanDir` (50-goroutine)
- Keycap emoji test, very long single-line test (1 MB), empty-file edge cases,
  `MaxFileBytes` boundary test, whitespace-only `SanitizeFile` change test,
  `TestScanDirContext`, and `example_test.go` examples for all new APIs

### Changed

- `scanDirCounted` accepts a `context.Context`; `ScanDir` passes `context.Background()`
  internally (no behavioral change for existing callers)
- `fix.go` updated to pass `context.Background()` to `scanDirCounted`
- `doc.go` expanded with godoc for all new APIs

### Fixed

- `SanitizeReport`: `EmojiRemoved` now reflects emoji actually removed
  (`CountEmoji(input) - CountEmoji(cleaned)`) rather than the raw input count;
  codepoints preserved by `AllowedEmojis` or `AllowedRanges` are no longer
  incorrectly counted as removed
- `SanitizeReader`: per-line scan buffer increased from the default 64 KiB to
  1 MiB (`sanitizeReaderMaxTokenSize`); lines exceeding 1 MiB return
  `bufio.ErrTooLong` (accommodates minified JSON, base64 payloads, long LLM
  output lines)
- `SanitizeJSON`: trailing non-whitespace bytes after the first JSON value now
  return an error instead of being silently ignored; inputs such as
  `{"a":1} trailing` or two concatenated JSON objects are rejected
- `cmd/demojify`: `writeJSON` now propagates encoder errors -- prints a
  diagnostic to stderr and exits 1 instead of discarding with `_ = enc.Encode`
- `scripts/hooks/pre-commit`: replaced `go run ...@latest` with local
  `./cmd/demojify` to eliminate network dependency and non-reproducibility;
  removed `exec` so both checks contribute to the final exit code; fixed
  typo "publshed" -> "published"
- `SanitizeFile` now skips binary files (NUL byte in first 512 bytes), matching
  `ScanFile`, `ScanDir`, and `ReplaceFile`; previously it could corrupt binary content
- `docs/design.md` "Scope boundaries" removed false claim about missing `io.Reader`/`io.Writer`
- `TestRepoAllDocsEmojiClean` and `TestRepoProductionFilesIdempotent` now include `docs/`

## [0.3.0] - 2026-03-05

### Added

- `FixDir(root string, cfg ScanConfig) (fixed, clean int, err error)` --
  write-side complement to `ScanDir`; walks a directory tree, applies the
  sanitization/replacement pipeline, and writes back every changed file in
  one call. Path-traversal protection validates every resolved write target
  stays within root
- `fix.go` / `fix_test.go` -- implementation and 9 subtests (basic fix,
  clean dir, SkipDirs, Extensions, replacements, idempotency, multiple
  files, bad root, path traversal rejection)
- `cmd/demojify` `-skip` flag -- comma-separated directory names to exclude
  in addition to the defaults (`.git`, `vendor`, `node_modules`); trailing
  slash auto-appended if omitted
- `cmd/demojify/main_test.go` `TestSkipFlag`, `TestSkipFlagWithTrailingSlash`

### Changed

- `.github/workflows/ci.yml` all actions pinned to immutable commit SHAs
  (`actions/checkout@34e11487...`, `actions/setup-go@40f1582b...`,
  `codecov/codecov-action@b9fd7d16...`) to prevent supply-chain tag mutation

### Security

- `fix.go` `isInsideDir` boundary check prevents path-traversal writes via
  `..` components in `Finding.Path`
- `SECURITY.md` expanded to document write-back function attack surface
  (`WriteFinding`, `SanitizeFile`, `ReplaceFile`, `FixDir`) with
  path-traversal guidance

## [0.2.4] - 2026-03-01

### Fixed

- `LICENSE` three word-level deviations from the official Apache-2.0 text
  corrected -- `"the Licensor"` vs canonical `"Licensor"` (two occurrences)
  and `"any notices"` vs canonical `"those notices"` (one occurrence); these
  caused `licensecheck`'s word-based LRE matcher to score below its
  detection threshold even with the APPENDIX present

### Added

- `repo_test.go` `TestRepoLicenseApache20ExactPhrases` -- pins 5 canonical
  phrases from the official Apache-2.0 text and detects known bad variants
  to prevent word-level drift from recurring

## [0.2.3] - 2026-03-01

### Fixed

- `LICENSE` missing `APPENDIX` section restored -- the canonical Apache-2.0
  text requires the `APPENDIX: How to apply the Apache License to your work.`
  boilerplate after `END OF TERMS AND CONDITIONS`; omitting it caused
  `github.com/google/licensecheck` (used by pkg.go.dev) to score coverage
  below its match threshold and report the license as unknown

### Added

- `repo_test.go` five license-hygiene tests pin every structural landmark
  that `licensecheck` requires for Apache-2.0 detection:
  - `TestRepoLicenseApache20Canonical` -- all 7 canonical text landmarks present
  - `TestRepoLicenseApache20SectionOrder` -- landmarks appear in correct order
  - `TestRepoLicenseFilename` -- file named exactly `LICENSE` in module root
  - `TestRepoLicenseNotEmpty` -- file is >= 10 KB (full canonical text required)
  - `TestRepoLicenseStartsOnLineOne` (pre-existing) retained and relocated

## [0.2.2] - 2026-03-01

### Fixed

- `LICENSE` leading blank line removed -- pkg.go.dev `licensecheck` scanner
  requires the Apache-2.0 header to begin on line 1 for automatic detection

## [0.2.1] - 2026-03-01

### Added

- `cmd/demojify` `-version` flag -- prints module version from build metadata
  and exits 0; reports `(devel)` for local `go run` builds

### Fixed

- `docs/cli.md` `-version` example corrected: `go run` always produces
  `(devel)` because the Go toolchain sets that marker directly in
  `debug.ReadBuildInfo()` for local source builds; added `go install` form
  to demonstrate real semver output
- `cmd/demojify/main_test.go` `TestMain` temp-dir leak: `defer os.RemoveAll`
  was bypassed by `os.Exit`; replaced with explicit cleanup before exit

## [0.2.0] - 2026-03-01

### Added

- Large-document benchmark suite (`benchmark_test.go`) covering all public
  APIs at 10 KB, 100 KB, and 1 MB with throughput and allocation reporting
- Binary-file detection in `ReplaceFile`, matching `ScanDir`/`ScanFile`
  behaviour
- `ExampleScanDir`, `ExampleScanFile`, `ExampleWriteFinding`,
  `ExampleSanitizeFile` in `example_test.go`
- Driver program (`docs/examples/driver/main.go`) exercising every major API
- Permission-denied and symlink test cases for `ScanDir` (skipped on Windows)

### Changed

- `isBinary` uses `bytes.IndexByte` instead of `bytes.ContainsRune` for
  faster byte-level detection
- `ScanDir` normalizes whitespace unconditionally when enabled, rather than
  gating on emoji presence
- CLI documentation (`docs/cli.md`) aligned with actual flag behavior

### Removed

- `Match.Emoji` deprecated field -- use `Match.Sequence`

## [0.1.0] - 2026-03-01

### Added

- `Demojify(string) string` -- strip all emoji codepoints
- `ContainsEmoji(string) bool` -- detect emoji presence
- `Normalize(string) string` -- collapse redundant whitespace, preserve
  leading indentation
- `Sanitize(string, Options) string` -- configurable pipeline (strip +
  normalize)
- `SanitizeFile(string, Options) (bool, error)` -- sanitize a file atomically
- `Options` / `DefaultOptions()` -- pipeline configuration
- `Options.AllowedRanges` -- preserve specific Unicode blocks during removal
- `Options.AllowedEmojis` -- preserve exact emoji strings during removal
- `Replace(string, map) string` -- substitute emoji with text equivalents
- `ReplaceFile(string, map) (int, error)` -- replace emoji in a file
- `ReplaceCount(string, map) (string, int)` -- replace and return count
- `FindAll(string) []string` -- distinct emoji sequences in text
- `FindAllMapped(string, map) []string` -- mapped keys found in text
- `DefaultReplacements() map[string]string` -- ~137 emoji-to-text entries
- `ScanConfig` / `DefaultScanConfig()` -- scanner configuration with
  directory/file exemptions and extension filters
- `ScanDir(ScanConfig) ([]Finding, error)` -- walk directory tree
- `ScanFile(string, Options) (*Finding, error)` -- check single file
- `FindMatchesInFile(string, map) ([]Match, error)` -- per-occurrence detail
- `Finding` / `Match` structs for scan results
- `WriteFinding(string, Finding) (bool, error)` -- atomic write-back
- `cmd/demojify` CLI with `-root`, `-fix`, `-sub`, `-normalize`, `-quiet`,
  `-exts` flags
- `repo_test.go` dogfooding tests (scan repo with own API)
- `example_test.go` with 17 runnable examples for pkg.go.dev
- Apache License 2.0

[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.4...v0.3.0
[0.2.4]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/nicholashoule/demojify-sanitize/releases/tag/v0.1.0
