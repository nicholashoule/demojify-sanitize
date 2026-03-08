---
applyTo: "**"
---

# General Project Instructions

## Project Overview

**demojify-sanitize** -- a dependency-free Go library that helps developers of web
applications and APIs audit, detect, and fix emoji clutter and redundant whitespace
before content reaches production. AI agents can import it to self-correct their
output, and applications can run it as a gate in their request or CI pipeline to
catch and fix every issue in one pass.

**Module path:** `github.com/nicholashoule/demojify-sanitize`

**Use cases:** Post-processing AI agent output in web apps and APIs, sanitizing
user-submitted content, running as a CI quality gate, cleaning Markdown and
documentation files, normalizing whitespace in automated text pipelines.

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21+ |
| Dependencies | None (pure stdlib) |
| Testing | Go `testing` package, table-driven tests |
| API Docs | pkg.go.dev (godoc) |
| CI | GitHub Actions |

## Package Structure

```
.
├── demojify.go # Demojify / ContainsEmoji / CountEmoji / BytesSaved / TechnicalSymbolRanges -- emoji removal
├── normalize.go # Normalize -- whitespace normalization
├── sanitize.go # Sanitize / SanitizeFile / SanitizeReport / SanitizeReader / SanitizeJSON / Options / DefaultOptions -- pipeline
├── scan.go # ScanConfig / ScanDir / ScanDirContext / ScanFile / FindMatchesInFile / Finding / Match -- scanner
├── replace.go # Replace / ReplaceFile / ReplaceCount / FindAll / FindAllMapped
├── replacements.go # DefaultReplacements -- built-in emoji-to-text map
├── write.go # WriteFinding -- atomic write-back for scan results
├── fix.go # FixDir / isInsideDir -- batch scan-and-fix
├── helpers.go # isBinary / sortByLenDesc / sortedKeys / statAndWrite / collapseInlineSpaces -- shared internals
├── doc.go # Package-level documentation
├── demojify_test.go # Tests for demojify.go
├── normalize_test.go # Tests for normalize.go
├── sanitize_test.go # Tests for sanitize.go
├── scan_test.go # Tests for scan.go
├── replace_test.go # Tests for replace.go
├── replacements_test.go # Tests for replacements.go
├── write_test.go # Tests for write.go
├── fix_test.go # Tests for fix.go
├── helpers_test.go # Shared test utilities (writeTempFile, writeTempDir)
├── example_test.go # Runnable pkg.go.dev examples
├── repo_test.go # Dogfooding tests -- validates repo with own API
├── go.mod # Module definition
├── Makefile # Development workflow
├── CHANGELOG.md # Release history
├── CONTRIBUTING.md # Contributor guide
├── SECURITY.md # Security policy
├── README.md # User-facing documentation
└── cmd/
 └── demojify/
 └── main.go # CLI tool for directory auditing and fixing
```

## Public API

| Symbol | File | Purpose |
|--------|------|---------|
| `Demojify(text string) string` | demojify.go | Remove all emoji codepoints |
| `ContainsEmoji(text string) bool` | demojify.go | Detect emoji presence |
| `CountEmoji(text string) int` | demojify.go | Count emoji codepoint occurrences |
| `BytesSaved(text string) int` | demojify.go | Bytes freed by emoji removal |
| `TechnicalSymbolRanges() []*unicode.RangeTable` | demojify.go | Pre-built ranges for LLM technical symbols |
| `Normalize(text string) string` | normalize.go | Collapse redundant whitespace |
| `Sanitize(text string, opts Options) string` | sanitize.go | Full configurable pipeline |
| `SanitizeFile(path string, opts Options) (bool, error)` | sanitize.go | Sanitize a file atomically |
| `SanitizeReport(text string, opts Options) SanitizeResult` | sanitize.go | Sanitize with structured metrics |
| `SanitizeResult` struct | sanitize.go | Cleaned text, EmojiRemoved, BytesSaved |
| `SanitizeReader(r io.Reader, w io.Writer, opts Options) error` | sanitize.go | Streaming line-by-line sanitization |
| `SanitizeJSON(data []byte, opts Options) ([]byte, error)` | sanitize.go | Sanitize JSON string values only |
| `Options` struct | sanitize.go | Pipeline configuration |
| `DefaultOptions() Options` | sanitize.go | All steps enabled |
| `Replace(text string, repl map[string]string) string` | replace.go | Substitute emoji with text equivalents |
| `ReplaceFile(path string, repl map[string]string) (int, error)` | replace.go | Replace emoji in a file atomically |
| `ReplaceCount(text string, repl map[string]string) (string, int)` | replace.go | Replace and return count |
| `FindAll(text string) []string` | replace.go | Distinct emoji sequences in text |
| `FindAllMapped(text string, repl map[string]string) []string` | replace.go | Mapped keys found in text |
| `FindMatchesInFile(path string, repl map[string]string) ([]Match, error)` | scan.go | Per-occurrence match detail |
| `DefaultReplacements() map[string]string` | replacements.go | Built-in emoji-to-text map |
| `ScanConfig` struct | scan.go | Scanner configuration |
| `DefaultScanConfig() ScanConfig` | scan.go | Sensible defaults, scans all file types |
| `ScanDir(cfg ScanConfig) ([]Finding, error)` | scan.go | Walk directory tree, return findings |
| `ScanDirContext(ctx context.Context, cfg ScanConfig) ([]Finding, error)` | scan.go | Context-aware scan with cancellation |
| `ScanFile(path string, opts Options) (*Finding, error)` | scan.go | Check a single file |
| `Finding` struct | scan.go | File path, emoji flag, original/cleaned content |
| `Match` struct | scan.go | Per-occurrence match with line, column, context |
| `WriteFinding(path string, f Finding) (bool, error)` | write.go | Write scan result back without re-reading |
| `FixDir(root string, cfg ScanConfig) (fixed, clean int, err error)` | fix.go | Scan and fix an entire directory tree in one call |

## Development Workflow

```bash
# Run all tests
make test

# Run tests with race detector
make race

# Run with coverage
make coverage

# Format code
make fmt

# Vet
make vet

# Lint (requires golangci-lint)
make lint

# Build check (library -- no binary)
make build
```

## Pre-Commit Checklist

```bash
make fmt # Format code
make vet # Run go vet
make test # All tests pass
make race # No data races
make coverage # Coverage >= 80%
```

## Git Commit Conventions

Format: `<type>(<scope>): <subject>`

| Type | Use |
|------|-----|
| `feat:` | New exported symbol or option |
| `fix:` | Bug fix |
| `docs:` | Documentation only |
| `style:` | Formatting (no logic change) |
| `refactor:` | Code restructuring |
| `test:` | Adding/updating tests |
| `chore:` | Maintenance, dependencies |
| `perf:` | Performance improvement |
| `ci:` | CI/CD changes |

**Scopes:** `(demojify)`, `(normalize)`, `(sanitize)`, `(docs)`, `(ci)`, `(config)`

**Rules:** imperative mood, no capitalization after type, no trailing period, 50-char limit.

## Agent Guidelines

1. **Type Safety** -- use Go's strong typing; avoid `interface{}` when concrete types work
2. **Cross-Platform** -- test and run correctly on Windows, Linux, and macOS
3. **No Dependencies** -- this library is intentionally dependency-free; do not add `go get` imports
4. **Performance** -- compile regexes at package init (`var re = regexp.MustCompile(...)`), never inside functions
5. **Error Handling** -- pure text-processing functions (`Demojify`, `ContainsEmoji`, `Normalize`, `Sanitize`) do not return errors; panics are only permitted in `MustCompile` at init time. Scanner functions (`ScanDir`, `ScanFile`) are the exception and must return `error` for filesystem failures.
6. **Testing** -- table-driven tests, `testing` package, >80% coverage target; update `example_test.go` when API changes
7. **No Emoji in Production Code** -- use text alternatives (`[PASS]`, `[FAIL]`, `WARNING:`) in all production source, comments, and output. Exception: literal emoji is permitted (and required) as test-input data inside `*_test.go` files
8. **Backward Compatibility** -- do not remove or rename exported symbols; add new `Options` fields instead

## Agent Token Optimization

- Use `grep_search` with `includePattern` for targeted searches
- Use `semantic_search` only for natural language matching
- Read 100+ lines per `read_file` call to capture full context
- Use `multi_replace_string_in_file` for batched edits
- Parallelize independent `read_file` and `grep_search` calls
- Reuse discovered file paths; don't re-search

| Task | Best Tool |
|------|-----------|
| Exact string search | `grep_search` |
| Function usage | `list_code_usages` |
| Concept search | `semantic_search` |
| File discovery | `file_search` |
| Type errors | `get_errors` |
| Multi-edit | `multi_replace_string_in_file` |

## Common Tasks

### Adding a new AI-clutter pattern
1. Add the pattern to `aiClutterRE` in `sanitize.go`
2. Add a test case in `sanitize_test.go` (both match and non-match)
3. Update the patterns table in `README.md`

### Adding a new Unicode emoji block
1. Add the range to `emojiRE` in `demojify.go`
2. Add a test case in `demojify_test.go`
3. Update the Unicode coverage table in `README.md`

### Adding a new Options field
1. Add the field to `Options` in `sanitize.go` with a doc comment
2. Set it to `true` in `DefaultOptions()` if it should be on by default
3. Add the step to `Sanitize()` in the correct pipeline order
4. Add test cases in `sanitize_test.go`
5. Update `README.md`

## Key References

- [library.instructions.md](library.instructions.md)
- [testing.instructions.md](testing.instructions.md)
- [README.md](../../README.md)
- [docs/design.md](../../docs/design.md)
- [emoji-prevention.instructions.md](emoji-prevention.instructions.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [CHANGELOG.md](../../CHANGELOG.md)