# CI Pipeline Integration

`demojify` is designed to run as a CI quality gate: audit the repository for
emoji, fail the build when any are found, and tell the author exactly which
file, line, and column to fix. This document shows the three integration
levels, from zero-config to fully embedded.

| Level | Mechanism | Best for |
|-------|-----------|----------|
| CLI gate | `go run .../cmd/demojify@<version>` as a pipeline step | Any repository, any language |
| Test gate | `ScanDir` inside a normal Go test | Go modules (no extra CI step at all) |
| JSON tooling | `-json` output consumed by scripts or bots | Custom reporting, PR annotations |

## CLI gate: GitHub Actions

A minimal audit-only gate. The job fails (exit `1`) when emoji are found and
prints each occurrence with file, line, and column:

```yaml
name: Emoji Gate

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  demojify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "stable"

      - name: Audit for emoji
        run: >
          go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.10.0
          -root . -exts .go,.md,.txt,.yaml,.yml,.json
```

Two deliberate choices in that step:

- **Pin the version** (`@v0.10.0`, not `@latest`) so the gate's behavior only
  changes when you choose to upgrade it.
- **Scope with `-exts`** to the text file types your repository authors.
  Compressed and binary assets are already skipped by the built-in
  `SkipExtensions` denylist, but an explicit allowlist keeps the audit fast
  and makes the gate's coverage obvious in the workflow file.

## CLI gate: GitLab CI

```yaml
emoji-gate:
  stage: test
  image: golang:1.24
  script:
    - >
      go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.10.0
      -root . -exts .go,.md -quiet
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
```

`-quiet` keeps the job log to the exit code; drop it to get the per-occurrence
report in the log when the gate fails.

## CLI gate: any CI system

Every CI system that can run a shell command can run the gate; only the exit
code matters (`0` clean, `1` findings, `2` usage error):

```sh
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.10.0 -root . -exts .go,.md
```

To auto-correct instead of failing, substitute emoji with text tokens and let
the pipeline commit or suggest the diff:

```sh
demojify -root . -sub          # rewrite in place, [PASS]/[FAIL]/[DEPLOY] tokens
git diff --exit-code           # fail if anything changed, with the fix as the diff
```

The audit-then-diff form gives reviewers a ready-made patch while still
failing the build, which is usually better in CI than silently pushing
auto-fix commits.

## Test gate: embed the audit in `go test`

For Go modules the cleanest integration needs no extra pipeline step at all:
add a test that walks the repository with the library's own scanner. Your
existing `go test ./...` CI job becomes the emoji gate.

```go
package myproject_test

import (
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// TestRepoEmojiClean fails when any production source or documentation
// file contains emoji, and names each offending file. Test files are
// exempt by default (DefaultScanConfig exempts *_test.go), so emoji used
// as test input data never trips the gate.
func TestRepoEmojiClean(t *testing.T) {
	cfg := demojify.DefaultScanConfig()
	cfg.Root = "."
	cfg.Extensions = []string{".go", ".md"}

	findings, err := demojify.ScanDir(cfg)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	for _, f := range findings {
		t.Errorf("%s contains emoji -- fix with: demojify -root . -sub", f.Path)
	}
}
```

Because Go excludes `*_test.go` files from downstream consumers, the gate
ships with your repository but never with your module. This repository
enforces its own hygiene exactly this way -- see `repo_test.go` for the
production version of the pattern, including an idempotency check and a
meta-test proving the test-file exemption is load-bearing.

For per-occurrence detail in test failures (line, column, context), set
`cfg.CollectMatches = true` and report `f.Matches` instead of just `f.Path`.

## JSON output for tooling

Pass `-json` for a stable machine-readable report -- useful for PR
annotation bots, dashboards, or custom gates that need per-occurrence data:

```sh
demojify -root . -json > findings.json
```

Example: turn findings into GitHub Actions error annotations, which surface
inline on the pull request diff:

```sh
go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.10.0 -root . -json |
  jq -r '.findings[] | .path as $p | .matches[] |
    "::error file=\($p),line=\(.line),col=\(.column)::emoji found (replace with \(.replacement // "removal"))"'
```

The JSON schema is documented in [cli.md](cli.md#json-output) and is a
stable API -- prefer it over parsing the text output.

## Choosing scope

- `-exts .go,.md` (or your project's text types) keeps the gate focused and
  fast; the built-in skip list already excludes binaries, archives, minified
  assets, and media.
- Test files are exempt by default because they legitimately contain emoji
  as input data. To scan them too, clear the exemption:
  `cfg.ExemptSuffixes = nil` (library) -- the binary/minified protection in
  `SkipExtensions` is retained.
- Add generated or vendored directories with `-skip dist,build` rather than
  widening `-exts`, so accidental emoji in generated text files still fail
  the gate everywhere else.

## See Also

- [cli.md](cli.md) -- full flag, exit-code, and JSON schema reference
- [git-hooks.md](git-hooks.md) -- the same gate as a local pre-commit hook,
  catching emoji before they ever reach CI
- [design.md](design.md) -- why the scanner skips what it skips
