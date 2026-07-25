# demojify CLI Reference

`demojify` is a command-line tool for auditing a directory tree for emoji and
Unicode pictographic characters. It reports every occurrence with file, line,
and column, and can optionally rewrite affected files in place -- stripping
emoji, substituting them with readable text tokens, or normalizing redundant
whitespace.

## Installation

Build from source:

```sh
go build -o demojify ./cmd/demojify/
```

Or install directly:

```sh
go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest
```

Or run without installing:

```sh
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest [flags]
```

## Synopsis

```
demojify [flags]
```

`demojify` has no subcommands; the operational mode is selected by flags.
Modes may be combined -- for example, `-sub -normalize` applies both
substitution and whitespace normalization in a single pass.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-root <dir>` | `.` | Directory to scan. |
| `-exts <.go,.md>` | all files | Comma-separated extensions to scan. Leading dot optional (`-exts go,md` works). |
| `-skip <dirs>` | _(none)_ | Comma-separated directory names to skip in addition to the defaults (`.git`, `vendor`, `node_modules`). Trailing slash auto-appended. |
| `-fix` | `false` | Rewrite affected files in place after reporting. |
| `-sub` | `false` | Substitute emoji with text tokens instead of stripping. Implies `-fix`. |
| `-normalize` | `false` | Collapse redundant whitespace in all scanned files. Implies `-fix`. |
| `-quiet` | `false` | Suppress all output except write errors (stderr); rely on exit code only. |
| `-json` | `false` | Output findings as JSON to stdout (overrides `-quiet`). |
| `-version` | `false` | Print the module version and exit. |

## Modes

### Audit (default)

Scan the directory tree and report all findings without writing anything.

```sh
demojify -root .
demojify -root . -exts .go,.md
demojify -root . -skip dist,build
```

Walks the tree, prints a per-occurrence report for every file containing
emoji, and exits `1` if any are found, `0` otherwise.

**Output columns:** `line N col M: "<sequence>" -> "<replacement>"`

### Fix (`-fix`)

Strip emoji in place after reporting.

```sh
demojify -root . -fix
demojify -root . -exts .go,.md -fix
```

Rewrites each affected file atomically (temp file + rename, original
permissions preserved), removing all emoji codepoints.

> **Note:** whitespace artifacts left by emoji removal (a double space where
> an emoji sat, trailing spaces at line end) are tidied only on the lines the
> removal actually changed. Untouched lines keep their exact whitespace, so
> gofmt tab alignment and column-aligned comments elsewhere in the file
> survive a fix pass. CRLF files keep their line endings.

### Substitute (`-sub`)

Replace emoji with text tokens instead of stripping them. Implies `-fix`.

```sh
demojify -root . -sub
demojify -root . -sub -json
```

Replaces each emoji with its text equivalent from `DefaultReplacements()`
(e.g., `[PASS]`, `[FAIL]`, `[DEPLOY]`, `->`). Emoji not present in the map
are stripped. Runs of adjacent identical emoji that map to the same token
collapse to a single token; literal text that happens to equal a token is
never altered.

### Normalize (`-normalize`)

Collapse redundant whitespace in every scanned file. Implies `-fix`.

```sh
demojify -root . -normalize
demojify -root . -sub -normalize
```

Collapses runs of inline spaces and tabs to a single space, trims trailing
whitespace, and limits consecutive blank lines. Leading indentation on each
line is preserved, so `-normalize` is safe for Markdown nested lists and
indented code blocks.

> **Note:** unlike `-fix`, normalization applies to the whole file, including
> lines with no emoji. Inline runs of spaces or tabs after the first
> non-whitespace character are collapsed, which destroys column-aligned
> comments and tabular formatting. Run gofmt after applying `-normalize` to
> Go source to restore comment alignment.

### Version (`-version`)

Print the version string and exit.

```sh
demojify -version
```

The version is read from the Go build info. A semver tag (e.g. `v0.10.0`)
is embedded only when the binary is installed from a published tagged
release (`go install ...@v0.10.0`). Builds from local source -- whether via
`go run`, `go build`, or `go install` without a version suffix -- report
`(devel)`.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No emoji found, or all findings fixed successfully |
| `1` | Emoji found without `-fix`, a write error occurred, or `-root` does not exist / is not a directory |
| `2` | Unknown flag (flag parse error) |

## Default Scan Behavior

The CLI uses `DefaultScanConfig()`, which skips:

- **Directories:** `.git/`, `vendor/`, `node_modules/` (plus any added via
  `-skip`).
- **File suffixes:** `*_test.go`.
- **Binary / minified / compressed / media extensions** -- never scanned or
  rewritten, skipped before the file is even opened:
  - Minified assets and source maps: `.min.js`, `.min.css`, `.js.map`,
    `.css.map`
  - Compressed / archive: `.gz`, `.tgz`, `.bz`, `.bz2`, `.xz`, `.zst`,
    `.zip`, `.tar`, `.7z`, `.br`
  - Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, `.ico`, `.bmp`,
    `.svgz`
  - Fonts: `.woff`, `.woff2`, `.ttf`, `.otf`, `.eot`
  - Media: `.mp3`, `.mp4`, `.wav`, `.mov`, `.webm`, `.ogg`
  - Documents / compiled: `.pdf`, `.exe`, `.dll`, `.so`, `.dylib`,
    `.wasm`, `.class`, `.jar`

All other file types are scanned unless `-exts` restricts them. Remaining
binary files are auto-detected (NUL byte sniff in the first 512 bytes) and
skipped. Files larger than 1 MiB are skipped.

## Text Output

Each finding is printed to stdout:

```
[WARN] path/to/file.go
  line 12 col 5: "<emoji>" -> "(stripped)"
  line 34 col 1: "<emoji>" -> "[PASS]"
  [PASS] fixed 2 occurrence(s)
```

When no findings exist:

```
[PASS] no emoji found
```

Write errors go to stderr (even with `-quiet`):

```
  [FAIL] write path/to/file.go: <error>
```

## JSON Output

Pass `-json` to receive machine-readable output. All output becomes a single
JSON object on stdout; the text format (`[WARN]`, `[PASS]`) is suppressed
entirely, and `-json` overrides `-quiet`.

The envelope has one key, `findings`:

```json
{
  "findings": [
    {
      "path": "internal/status.go",
      "hasEmoji": true,
      "matches": [
        {
          "sequence": "<raw UTF-8 emoji>",
          "replacement": "[PASS]",
          "line": 12,
          "column": 5,
          "context": "// build result <raw UTF-8 emoji>"
        }
      ],
      "fixed": {
        "success": true,
        "count": 1
      }
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Forward-slash relative file path |
| `hasEmoji` | bool | Whether the file contains emoji |
| `matches` | array | Per-occurrence detail (omitted when empty) |
| `matches[].sequence` | string | Matched codepoint sequence (raw UTF-8) |
| `matches[].replacement` | string | Mapped substitute or empty string |
| `matches[].line` | int | 1-based line number |
| `matches[].column` | int | 0-based byte offset within the line |
| `matches[].context` | string | Full source line containing the match |
| `fixed` | object | Present only when `-fix` or `-sub` is set |
| `fixed.success` | bool | Whether the write-back succeeded |
| `fixed.count` | int | Number of occurrences fixed |
| `fixed.error` | string | Error message (omitted on success) |

When no findings exist the output is `{"findings":[]}` with exit code `0`.

**Stability:** the JSON output format is a stable, machine-readable API.
Downstream consumers should prefer `-json` over parsing the text format.

## Examples

```sh
# Audit the current tree (no writes); exit 1 if emoji found
demojify -root .

# Strip emoji in place
demojify -root . -fix

# Substitute emoji with text tokens
demojify -root . -sub

# Substitute and normalize whitespace in one pass
demojify -root . -sub -normalize

# Restrict to Go and Markdown files (leading dot optional)
demojify -root . -exts .go,.md

# Skip additional directories
demojify -root . -skip dist,build

# CI gate: quiet mode, exit code only
demojify -root . -quiet

# Machine-readable JSON output for tooling
demojify -root . -json

# Fix, then verify clean
demojify -root . -sub && demojify -root . -exts .go,.md

# Print version
demojify -version
```

## Relationship to the Library API

The CLI is a thin wrapper around the library. The equivalent library calls:

| CLI flag combination | Library call |
|----------------------|--------------|
| `-version` | `runtime/debug.ReadBuildInfo()` (no library call) |
| (audit only) | `ScanDir(DefaultScanConfig())` |
| `-fix` | `ScanDir(cfg)` + `WriteFinding(path, f)` |
| `-sub` | `ScanDir(cfg)` with `cfg.Replacements = DefaultReplacements()` + `WriteFinding(path, f)` |
| `-sub -normalize` | as `-sub`, plus `cfg.Options.NormalizeWhitespace = true` |
| `-skip dist,build` | `cfg.SkipDirs = append(cfg.SkipDirs, "dist/", "build/")` |
| `-json` | Same as audit; output wrapped in JSON envelope with `Finding`/`Match` fields |

## See Also

- [ci.md](ci.md) -- CI pipeline integration: GitHub Actions, GitLab CI, and test-based gates
- [git-hooks.md](git-hooks.md) -- pre-commit hook integration that runs these checks automatically
- [replacements.md](replacements.md) -- full `DefaultReplacements()` token reference
- [design.md](design.md) -- architecture and design decisions
