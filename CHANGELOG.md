# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2026-05-16

### Added

- `scan.go` (`ScanConfig.SkipExtensions`, `DefaultScanConfig`): new
  `SkipExtensions` field listing binary, minified, compressed, and media
  file suffixes that are never scanned or rewritten. `DefaultScanConfig`
  pre-populates it with a comprehensive set — `.min.js`, `.min.css`,
  `.js.map`, `.css.map`; `.gz`/`.tgz`/`.bz`/`.bz2`/`.xz`/`.zst`/`.zip`/
  `.tar`/`.7z`/`.br`; common image/font/media suffixes; and
  `.pdf`/`.exe`/`.dll`/`.so`/`.dylib`/`.wasm`/`.class`/`.jar`. Matched
  files are skipped before the file is opened, eliminating false positives
  in minified web assets and avoiding I/O the audit can never act on. The
  field is independent of `ExemptSuffixes`, so clearing `ExemptSuffixes`
  to scan `*_test.go` files retains binary/minified protection
- `cmd/demojify/main.go`: custom `flag.Usage` so `demojify -h` prints a
  synopsis, the flag-selected operational modes, exit codes, and worked
  examples instead of the bare flag dump (which also leaked the absolute
  binary path on Windows)
- `cmd/demojify/main.go`: package doc gained `# CLI Exit Codes` and
  `# CLI JSON Output` sections so the pkg.go.dev page is the full CLI
  reference the library `doc.go` cross-reference already promised
- `Makefile`: `help` target (lists every target) and `pre-commit` target.
  `pre-commit` invokes the canonical cross-platform gate
  (`go run scripts/hooks/pre-commit.go`: gofmt, vet, golangci-lint, full
  test suite) directly, so it works whether `make` runs recipes via `sh`
  or `cmd.exe` — unlike the POSIX-shell `fmt-check` target, which fails
  under a `cmd.exe`-based `make` on Windows

### Changed

- Pinned tool versions are now consistent across the operative hook and
  every example. `scripts/hooks/pre-commit` and the README / `docs/git-hooks.md`
  hook snippets all reference `demojify-sanitize/cmd/demojify@v0.8.0` (was a
  mix of `@v0.7.1`/`@v0.7.3`) and `repogov/cmd/repogov@v0.7.0` (docs were a
  stale `@v0.3.0`); `.github/rules/emoji-prevention.md` install example
  bumped `@v0.7.0` → `@v0.8.0`
- `docs/git-hooks.md`: the Windows PowerShell hook example now passes
  `-exts .go,.md` to `demojify`, matching the `sh` examples so all hook
  snippets share the binary-false-positive scoping introduced in #28
- `Makefile`: the `fmt-check`, `hooks`, and `clean` targets now delegate to
  cross-platform Go helpers (`scripts/fmtcheck.go`, `scripts/installhooks.go`,
  `scripts/clean.go`) instead of POSIX `gofmt`/`cp`/`chmod`/`rm` shell
  recipes, so they work whether `make` runs recipes via `sh` or `cmd.exe`
  on Windows. `fmt-check` remains report-only (never runs `gofmt -w`)

### Fixed

- `demojify_test.go` (`TestDemojifyPreservesNonLatinScripts`): added regression
  test asserting that codepoints from non-Latin Unicode scripts — Devanagari
  (`राज`, U+0930/U+093E/U+091C and full sentences), Arabic, Hebrew, CJK, Cyrillic,
  Thai, and extended Latin with diacritics — pass through `Demojify` unchanged
  and that `ContainsEmoji` returns `false` for each. Addresses the false-positive
  report in issue #26: investigation confirmed the library was already correct
  (these scripts do not overlap any range in `emojiRE`); the test guards against
  future regressions if the covered Unicode ranges are ever extended
- `scan_dir_test.go` (`TestDefaultScanConfigSkipsBinaryAndMinified`,
  `TestDefaultScanConfig`): added a brotli regression guard. A `tailwind.css.br`
  blob whose bytes have no early NUL (so the binary NUL-sniff would not skip it)
  yet decode `U+27A2` (a rightwards-arrowhead glyph inside `emojiRE`'s
  U+2600–U+27BF range) is asserted
  to be skipped via the default `SkipExtensions` `.br` entry, and `.br` is now
  a required member of the spot-checked default skip set. Addresses a
  downstream false-positive report where a precompressed asset was scanned;
  investigation confirmed `DefaultScanConfig` already skips `.br` before the
  file is opened — the guard prevents that protection from regressing

## [0.8.0] - 2026-03-23

### Added

- `replacements.go` (`DefaultReplacements`): expanded from ~230 entries and 18
  categories to ~290 entries and 20 categories; doc comment updated accordingly.
  New and extended entries:
  - **Calendar** (new category): `\U0001F4C5`/`\U0001F4C6` → `[DATE]`;
    `\U0001F5D3`/`\U0001F5D3\uFE0F` → `[CALENDAR]`
  - **Scissors** (new category): `\u2702`/`\u2702\uFE0F` → `[REMOVED]`
  - **Deprecated** (new category): `\U0001FAA6` → `[DEPRECATED]`;
    `\U0001F4DB` → `[DEPRECATED]`
  - **Flags** (new category, 27 entries): single-codepoint flag emoji
    (`\U0001F6A9`, `\U0001F3F3`, `\U0001F3F4`, `\U0001F38C`) and VS-16 variants
    → `[FLAG]`; ZWJ sequences for flag → `[FLAG]`; tag-sequence flags for England, Scotland, and Wales → `[FLAG]`;
    17 regional-indicator pairs (US, GB, DE, FR, JP, CA, AU, BR, IN, CN, RU,
    KR, MX, NG, ZA, SA, AE) → `[FLAG]`
  - **Media controls**: `\u23ED`/`\u23ED\uFE0F` → `[SKIP]`;
    `\u23EE`/`\u23EE\uFE0F` → `[PREV]`
  - **Community/status**: `\U0001F53C` → `[UP]`; `\U0001F53D` → `[DOWN]`;
    `\U0001F446`/`\U0001F447`/`\U0001F448` → `[SEE]`;
    `\U0001F6A5`/`\U0001F6A6` → `[STATUS]`
  - **Platform**: `\U0001F34E` → `[MACOS]`; `\U0001FA9F` → `[WINDOWS]`
- Expanded demojify end-to-end fixtures: comprehensive emoji test corpus
  (~870 lines, 34,249 bytes) covering all `DefaultReplacements` entries, ZWJ
  sequences, variation selectors, skin tones, keycap sequences, subdivision
  flags, regional indicators, and Unicode 14/15/16 additions; verified
  against both `-sub` (2,044 substitutions) and `-fix -normalize` (2,135
  occurrences stripped) passes

### Fixed

- `repo_test.go` (`TestRepoAllDocsEmojiClean`, `TestRepoProductionFilesIdempotent`):
  added `"tmp/"` to `cfg.SkipDirs` to exempt `cmd/demojify/tmp/` from the
  repo-wide emoji hygiene scans; `tmp/` holds intentional emoji-laden test
  fixtures analogous to the existing `_test.go` suffix exemption

## Older releases

Entries for v0.7.3 and earlier were trimmed to keep this file within the
repository's documentation line limits. The complete history is preserved
in the per-tag [GitHub releases] and in `git log`; each version's diff is
reachable from the compare links below.

[GitHub releases]: https://github.com/nicholashoule/demojify-sanitize/releases

[Unreleased]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/nicholashoule/demojify-sanitize/compare/v0.7.3...v0.8.0
