# Emoji Prevention

This document describes how `demojify-sanitize` is used to enforce an emoji-free
codebase for this repository itself -- dogfooding the library against its own source.

## Why emoji-free source?

- **Token efficiency** -- emoji in documentation wastes AI context window tokens
  (each emoji encodes to multiple UTF-8 bytes and typically multiple LLM tokens)
- **Terminal portability** -- not all CI runners and log viewers render emoji correctly
- **Diff readability** -- emoji changes add noise to git diffs
- **Consistency** -- text alternatives (`[PASS]`, `[FAIL]`, `WARNING:`) are grep-able
  and work in every environment

## Text alternatives

| Instead of | Write |
|------------|-------|
| checkmark emoji | `[PASS]` or `[OK]` |
| cross/X emoji | `[FAIL]` or `[ERROR]` |
| warning emoji | `WARNING:` |
| info/note emoji | `[INFO]` or `NOTE:` |
| lightbulb emoji | `TIP:` |
| rocket emoji | `Deployment` or `Released` |
| chart emoji | `Report` or `Metrics` |
| star emoji | `[FEATURED]` |
| lock emoji | `Security` |

## Detection: using the module

[`ContainsEmoji`](../demojify.go) detects emoji in any string.
[`Sanitize`](../sanitize.go) removes them along with AI clutter and whitespace.

```go
import demojify "github.com/nicholashoule/demojify-sanitize"

data, _ := os.ReadFile("README.md")
if demojify.ContainsEmoji(string(data)) {
    clean := demojify.Sanitize(string(data), demojify.DefaultOptions())
    _ = os.WriteFile("README.md", []byte(clean), 0o644)
}
```

## Automated enforcement: dogfooding tests

[`repo_test.go`](../repo_test.go) runs four tests on every `go test` invocation,
using this module's own API against the repository's files:

| Test | What it checks |
|------|---------------|
| `TestRepoAllSourceFilesEmojiClean` | Every `.go` file contains no literal emoji |
| `TestRepoAllDocsEmojiClean` | Every `.md` file (except README.md) contains no emoji, including all `.github/` files |
| `TestRepoAllFilesIdempotent` | `Sanitize` (emoji + AI clutter removal) on every file is a no-op -- files are already clean |
| `TestRepoAgentOutputRemediation` | Proves the module detects and fully cleans realistic rogue AI output, and that the result is idempotent |

Files are discovered dynamically via `filepath.WalkDir` -- no hardcoded lists.
Adding a new file to the repo automatically brings it under enforcement.
The idempotent test is the strongest guarantee: if any file drifts, `Sanitize`
modifies it and the test fails, naming the offending file and the fix command.

### Running the checks

```bash
# Run all tests including repo self-checks
make test

# Run only the dogfooding tests
go test -run TestRepo ./...
```

## Unicode coverage

The following ranges are detected by `ContainsEmoji` (defined in [`demojify.go`](../demojify.go)):

| Range | Description |
|-------|-------------|
| U+231A–U+231B | Watch, Hourglass |
| U+23CF, U+23E9–U+23F3, U+23F8–U+23FA | Media controls |
| U+24C2 | Circled M |
| U+25AA–U+25AB, U+25B6, U+25C0, U+25FB–U+25FE | Geometric shapes |
| U+2600–U+27BF | Miscellaneous Symbols + Dingbats |
| U+2934–U+2935, U+2B05–U+2B07 | Curved/directional arrows |
| U+2B1B–U+2B1C, U+2B50, U+2B55 | Large squares, star, circle |
| U+3030, U+303D, U+3297, U+3299 | CJK/wavy dash symbols |
| U+1F000–U+1FAFF | All supplementary emoji blocks |
| U+200D | Zero Width Joiner |
| U+20E3 | Combining Enclosing Keycap |
| U+FE00–U+FE0F | Variation Selectors 1–16 |

Intentionally **not** detected: `©` (U+00A9), `®` (U+00AE), `™` (U+2122),
mathematical arrows (U+2190–U+2193), and all language scripts.

## What is not checked

`README.md` is intentionally excluded from `TestRepoDocsEmojiClean` because it
contains emoji inside code-fence examples to illustrate what the API removes.
Those appear as string literals in example output, not as decorative content.

If you need to add emoji to other documentation for a legitimate illustrative
purpose, use Unicode escape sequences in source (e.g. `\U0001F680`) rather than
literal characters, so the file itself passes `ContainsEmoji`.

## In code: escaping instead of literals

Source files that must reference emoji (e.g. test inputs) should use Go Unicode
escape syntax so the file bytes are ASCII-clean:

```go
// Correct -- source is emoji-free, behavior is tested
demojify.Demojify("\U0001F680 Deploy complete!")

// Incorrect -- embeds emoji bytes in source (do not copy this pattern)
demojify.Demojify("[U+1F680 rocket emoji] Deploy complete!")
```

See [`example_test.go`](../example_test.go) for the applied convention.
