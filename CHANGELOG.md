# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.3] - 2026-03-23

### Fixed

- `replace.go` (`collapseRepeatedTokens`): added `len(v) < 4` guard to skip
  single-character and short ASCII replacement values (e.g. `"/"` from U+2797,
  `"-"` from U+2796, `"*"` from U+2022/U+25CF, `"o"` from U+25CB) that appear
  legitimately throughout source code and documentation. Previously, running
  `Replace` with `DefaultReplacements()` would collapse `//` → `/` (breaking
  every URL and Go comment), `**` → `*` (converting Markdown bold to italic),
  `--` → `-` (stripping CLI flag docs), and `oo` → `o` (corrupting words like
  `root` and `bool`). Only label-like tokens of 4+ characters (`[FAIL]`,
  `WARNING`, `[DEPLOY]`, etc.) are now eligible for deduplication — the
  intended behavior when two adjacent identical emoji both produce the same
  substitution token
- `scan.go` (`scanDirCounted`): CRLF (`\r\n`) line endings are now restored
  after the internal LF-normalization pass that precedes inline-space cleanup.
  Previously, any CRLF file containing emoji had its line endings silently
  converted to LF during emoji removal — even when `NormalizeWhitespace` was
  false — creating noisy git diffs unrelated to emoji changes. The fix detects
  `hasCRLF` before the cleanup block and re-applies `\r\n` in the output when
  the original file used Windows line endings

### Added

- `replace_test.go`: `TestReplaceDefaultReplacementsPreservesASCII` — regression
  test asserting that `//`, `**`, `--`, `oo` (in `root`, `bool`, `tool`), `++`,
  and full URL strings are never collapsed when using `Replace` with
  `DefaultReplacements()`
- `replace_test.go`: `TestReplaceLongTokensStillCollapse` — complementary test
  verifying that adjacent identical emoji (warning sign x2, check mark x3, rocket x2) still collapse
  to a single label token after the `collapseRepeatedTokens` fix
- `scan_dir_test.go`: `TestScanDirPreservesCRLF` — regression test verifying
  that emoji is removed and CRLF line endings are preserved in the `Cleaned`
  output when `NormalizeWhitespace` is false
- `scan_dir_test.go`: `TestScanDirPreservesCRLFCleanFile` — verifies that a
  CRLF file with no emoji produces no finding (clean files are never reported)
- `demojify_test.go`: `TestDemojifyPreservesLegalSymbols` — documents and
  asserts that © (U+00A9), ® (U+00AE), and ™ (U+2122) pass through `Demojify`
  unchanged; these codepoints carry the Unicode emoji property in some contexts
  but are standard legal/documentation characters that must not be stripped
- `.github/workflows/ci.yml`: `os-compat` job — dedicated per-OS compatibility
  job running on `ubuntu-latest`, `macos-latest`, and `windows-latest` with
  `fail-fast: false`; each OS-sensitive concern is a named step so regressions
  are surfaced by category in the PR status panel rather than buried across 9
  matrix cells:
  - `CRLF line-ending preservation` (`TestScanDirPreservesCRLF`)
  - `Atomic write and file permissions` (`TestWriteFinding`, `TestReplaceFile`)
  - `Binary file detection` (`TestScanDirSkipsBinaryFiles`, `...NulAfterSniffSize`)
  - `File read protection (chmod)` (`TestScanDirUnreadableFile`)
  - `Symlink handling` (`TestScanDirSymlink`, `TestFixDir/rejects_symlink`)
  - `Path traversal protection` (`TestFixDir/does_not_write_outside_root`)
  - `ASCII preservation regression` (`TestReplaceDefaultReplacementsPreservesASCII`,
    `TestReplaceLongTokensStillCollapse`)

## [0.7.2] - 2026-03-22

### Fixed

- `scan_dir_test.go`: removed mismatched `TestScanFileEmptyFile` doc comment
  (now replaced by a correct `TestScanDirEmptyFile` comment adjacent to that
  function); removed orphaned `isWindows` comment (canonical comment lives in
  `helpers_test.go`); removed stale `TestScanDirNormalizeUnconditional` comment
  (relocated to `scan_matches_test.go` where the function is defined)
- `scan_matches_test.go`: `TestScanDirNormalizeUnconditional` doc comment added
  directly above the function
- `sanitize_io_test.go`: replaced misattributed `TestSanitizeAgentOutputRemediation`
  comment with a correct `TestSanitizeReport` comment
- `sanitize_test.go`: `TestSanitizeAgentOutputRemediation` doc comment added
  directly above the function
- `CHANGELOG.md` [0.7.1] entry: number formatted as `1,259` (was `1 259`)

## [0.7.1] - 2026-03-22

### Changed

- Test suite restructured: four monolithic test files split into 14
  contextually focused files, reducing the largest file from 1,259 to 634
  lines; all 424 tests preserved with no behavioral change
  - `scan_test.go` (deleted) split into `scan_dir_test.go`,
    `scan_matches_test.go`, and `scan_file_test.go`
  - `sanitize_test.go` (trimmed) + new `sanitize_io_test.go`
    (`SanitizeFile`, `SanitizeReport`, `SanitizeReader`, `SanitizeJSON`)
  - `cmd/demojify/main_test.go` (trimmed to infrastructure only) + new
    `cli_audit_test.go`, `cli_flags_test.go`, `cli_json_test.go`,
    `cli_space_test.go`
  - `replace_test.go` (trimmed) + new `replace_file_test.go`
    (`ReplaceFile`, concurrent safety)
  - `helpers_test.go`: `isWindows()` helper added (moved from
    `scan_test.go`) and made available across the whole test package

## [0.7.0] - 2026-03-12

### Added

- `config.go`: `LimitConfig` struct, `DefaultLimitConfig() LimitConfig`,
  `ResolveLimit(cfg LimitConfig, path string) int`, and
  `DefaultLineLimit` constant -- per-file line limit configuration with a
  file-specific override map; `DefaultLimitConfig` sets a 500-line default and
  caps `.claude/CLAUDE.md` at 50 lines
- `config_test.go`: black-box tests in `package demojify_test` covering exported API
  (`TestDefaultLimitConfig`, `TestDefaultLineLimit`, `TestLimitConfig_ZeroDefaultFallback`,
  `TestLimitConfig_FileOverride`, `TestResolveLimit`)
- `replace.go`: `distinctValues()` -- deduplicated, length-sorted slice of
  non-empty replacement values; `collapseRepeatedTokens()` -- collapses runs
  of the same token (space-separated or directly concatenated) to one occurrence
- `cmd/demojify/main_test.go`: `TestFixCollapseSpacesAfterEmojiRemoval` (8
  sub-tests), `TestSubCollapseSpacesAfterSubstitution` (8 sub-tests), and
  `TestFixSubIdempotentAfterSpaceCleanup` -- end-to-end CLI coverage of the
  space-collapse and repeated-token-collapse behaviour for `-fix` and `-sub`
- `docs/design.md`: "Per-file line limit configuration" section explaining the
  rationale for `LimitConfig` and its sentinel-zero behavior
- `README.md`: "Line limit configuration" API subsection documenting
  `LimitConfig`, `DefaultLimitConfig`, and `DefaultLineLimit`
- `.github/rules/`: 9 scoped Copilot instruction files (`backend.md`,
  `codereview.md`, `emoji-prevention.md`, `frontend.md`, `general.md`,
  `governance.md`, `library.md`, `repo.md`, `testing.md`) replacing
  `.github/instructions/`

### Changed

- `replace.go` (`applyReplacer`): consecutive identical substitution tokens
  produced by adjacent repeated emoji are now collapsed to a single token
  (e.g. two adjacent warning-sign emoji → `"WARNING"` instead of `"WARNINGWARNING"`); both
  space-separated runs and direct concatenations are handled
- `scan.go` (`scanDirCounted`): when `NormalizeWhitespace` is not requested,
  inline double-spaces and trailing spaces left as artifacts of emoji removal
  or substitution are now automatically collapsed via `collapseInlineSpaces`
  and `trailingSpaceRE`; callers no longer need `-normalize` for a
  whitespace-clean result
- `scan_test.go`: `TestScanDirReplacementsUnmappedEmojiStripped` expected
  output updated from `"[PASS] done  launch\n"` to `"[PASS] done launch\n"`
  to match the new space-collapse behaviour
- `cmd/demojify/main.go`: simplified write path -- `WriteFinding` is now always
  used to write cleaned content regardless of the `-sub`/`-normalize`
  combination; removed the separate `ReplaceFile` branch for `-sub` without
  `-normalize`
- `.github/copilot-instructions.md`, `AGENTS.md`: updated to reference
  `.github/rules/` for scoped instructions

### Removed

- `.github/ISSUE_TEMPLATE/bug_report.md`, `documentation.md`,
  `feature_request.md` -- replaced by the repository's default issue templates
- `.github/instructions/codereview.instructions.md`,
  `emoji-prevention.instructions.md`, `general.instructions.md`,
  `library.instructions.md`, `testing.instructions.md` -- migrated to
  `.github/rules/` with updated `applyTo` frontmatter

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

[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.3...HEAD
[0.7.3]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.6.0...v0.7.0
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
