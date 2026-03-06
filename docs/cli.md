# CLI Reference: demojify

The `cmd/demojify` tool audits a directory tree for emoji, reports every
occurrence with file, line, and column, and optionally rewrites affected files.

## Installation

```bash
# Run directly with go run
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify [flags]

# Or build a binary
go build -o demojify ./cmd/demojify
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-root <dir>` | `.` | Directory to scan |
| `-exts <.go,.md>` | all files | Comma-separated extensions to scan |
| `-fix` | false | Rewrite affected files in place after reporting |
| `-sub` | false | Substitute emoji with text tokens instead of stripping; implies `-fix` |
| `-normalize` | false | Collapse redundant whitespace in all scanned files; implies `-fix` |
| `-quiet` | false | Suppress all output; exit code only (0 = clean, 1 = findings/errors) |
| `-skip <dirs>` | none | Comma-separated directory names to skip in addition to defaults; trailing slash auto-appended |
| `-version` | false | Print the module version and exit 0 |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No emoji found, or all findings fixed successfully |
| `1` | Emoji found and `-fix` not specified, or a write error occurred |

The program exits immediately with a fatal log message (non-zero) if `-root`
does not exist or is not a directory.

## Output Format

Each finding is printed to stdout:

```
[WARN] path/to/file.go
  line 12 col 5: "[EMOJI]" -> "(stripped)"
  line 34 col 1: "[EMOJI]" -> "[PASS]"
  [PASS] fixed 2 occurrence(s)
```

When no findings exist:

```
[PASS] no emoji found
```

Write errors go to stderr:

```
  [FAIL] write path/to/file.go: <error>
```

## Default Scan Behavior

Uses `DefaultScanConfig()`, which skips:
- Directories: `.git/`, `vendor/`, `node_modules/`
- File suffixes: `_test.go`

All file types are scanned unless `-exts` restricts them.
Binary files are auto-detected (NUL byte sniff) and skipped.
Files larger than 1 MiB are skipped.

## Examples

### Print version

```bash
go run ./cmd/demojify -version
# demojify (devel)
```

`go run` always builds from local source, so `debug.ReadBuildInfo()` sets the
version to `(devel)` directly. An empty-string fallback in `cliVersion()` also
produces `(devel)` as a defensive guard for unusual non-module build contexts.

To see a real semver tag, install a published version and invoke the binary:

```bash
go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest
demojify -version
# demojify vX.Y.Z
```

### Audit only (no writes)

```bash
go run ./cmd/demojify -root .
```

Reports all emoji found; exits 1 if any are found.

### Strip emoji in place

```bash
go run ./cmd/demojify -root . -fix
```

Rewrites affected files, stripping all emoji codepoints.

### Substitute emoji with text tokens

```bash
go run ./cmd/demojify -root . -sub
```

Replaces each emoji with its text equivalent from `DefaultReplacements()`
(e.g., `[PASS]`, `[FAIL]`, `[DEPLOY]`, `->`). Residual unmapped emoji are
stripped. Implies `-fix`.

### Substitute and normalize whitespace

```bash
go run ./cmd/demojify -root . -sub -normalize
```

After substitution, collapses multiple consecutive spaces on each line to a
single space and removes trailing whitespace. Useful when the original content
had emoji surrounded by spaces that would otherwise leave double spaces behind.

Leading indentation on each line is preserved, so `-normalize` is safe for
Markdown nested lists and indented code blocks. However, inline runs of
multiple spaces or tabs after the first non-whitespace character are collapsed
to a single space, which will destroy column-aligned comments and tabular
formatting. Run gofmt after applying `-normalize` to Go source to restore
comment alignment.

### Audit Markdown and Go files only

```bash
go run ./cmd/demojify -root . -exts .md,.go
```

The leading dot is optional -- `-exts go,md` is equivalent to `-exts .go,.md`.

### Skip specific directories

```bash
go run ./cmd/demojify -root . -skip dist,build
```

Excludes `dist/` and `build/` (in addition to the defaults: `.git/`, `vendor/`,
`node_modules/`). A trailing slash is auto-appended if omitted.

### CI gate -- fail the build if emoji are found

```bash
go run ./cmd/demojify -root . -exts .go,.md
echo "Exit: $?"
```

Exits 0 (pass) or 1 (fail). Combine with `-sub` to auto-correct instead.

Use `-quiet` in CI pipelines where only the exit code matters:

```bash
go run ./cmd/demojify -root . -quiet
```

### Fix, then verify clean

```bash
go run ./cmd/demojify -root . -sub && \
go run ./cmd/demojify -root . -exts .go,.md
```

## Relationship to the Library API

The CLI is a thin wrapper around the library. The equivalent library calls are:

| CLI flag combination | Library call |
|----------------------|--------------|
| `-version` | `runtime/debug.ReadBuildInfo()` (no library call) |
| (audit only) | `ScanDir(DefaultScanConfig())` |
| `-fix` | `ScanDir(cfg)` + `WriteFinding(path, f)` |
| `-sub` | `ScanDir(cfg)` with `cfg.Replacements = DefaultReplacements()` + `ReplaceFile(path, repl)` |
| `-sub -normalize` | `ScanDir(cfg)` with `cfg.Options.NormalizeWhitespace = true` + `WriteFinding(path, f)` |
| `-skip dist,build` | `cfg.SkipDirs = append(cfg.SkipDirs, "dist/", "build/")` |
