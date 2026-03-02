# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3] - 2026-03-01

### Fixed

- `LICENSE` missing `APPENDIX` section restored -- the canonical Apache-2.0
  text requires the `APPENDIX: How to apply the Apache License to your work.`
  boilerplate after `END OF TERMS AND CONDITIONS`; omitting it caused
  `github.com/google/licensecheck` (used by pkg.go.dev) to score coverage
  below its match threshold and report the license as unknown
- `LICENSE` three word-level deviations from the official Apache-2.0 text
  corrected -- `"the Licensor"` vs canonical `"Licensor"` (two occurrences)
  and `"any notices"` vs canonical `"those notices"` (one occurrence); these
  caused `licensecheck`'s word-based LRE matcher to score below its
  detection threshold even with the APPENDIX present

### Added

- `repo_test.go` six license-hygiene tests pin every structural landmark
  and exact phrase that `licensecheck` requires for Apache-2.0 detection:
  - `TestRepoLicenseApache20Canonical` -- all 7 canonical text landmarks present
  - `TestRepoLicenseApache20SectionOrder` -- landmarks appear in correct order
  - `TestRepoLicenseApache20ExactPhrases` -- word-level canonical phrase checks
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

[0.2.3]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/nicholashoule/demojify-sanitize/releases/tag/v0.1.0
