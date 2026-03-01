# Contributing to demojify-sanitize

Thank you for your interest in contributing. This is a focused, dependency-free
Go library for text sanitization. Contributions should remain within that scope.

## Before You Start

- Search [existing issues](https://github.com/nicholashoule/demojify-sanitize/issues)
 to avoid duplicate work.
- For non-trivial changes, open an issue first to discuss the approach.
- All contributions must maintain the library's zero-dependency guarantee.

## Development Setup

```bash
git clone https://github.com/nicholashoule/demojify-sanitize.git
cd demojify-sanitize
go test ./... # confirm baseline passes
```

No additional tooling is required. Optional:
- `golangci-lint` for `make lint` ([installation](https://golangci-lint.run/usage/install/))

> **Windows note:** `make race` requires CGO and a C compiler (`gcc`).
> Install gcc via [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/),
> then set `CGO_ENABLED=1`. If gcc is unavailable, skip `make race` locally --
> the race detector runs automatically in CI (Ubuntu) on every push.

## Making Changes

1. Fork the repository and create a feature branch.
2. Follow the pre-commit checklist (`make fmt vet test race`).
3. Write or update tests. Coverage must remain >=80%.
4. Update `README.md` and `CHANGELOG.md` if the public API or behavior changes.
5. Open a pull request against `main`.

## Code Standards

- **No external dependencies.** This is a hard requirement.
- **Compiled regexes at package level.** All `regexp.MustCompile` calls must be
 `var` declarations, never inside functions.
- **Table-driven tests.** See existing `*_test.go` files for the pattern.
- **`gofmt -s` formatting.** Run `make fmt` before committing.
- **Godoc comments on all exported symbols.**
- **No emoji in production source, comments, or output.** Use `[PASS]`, `[FAIL]`,
 `WARNING:` etc. The sole exception is `*_test.go` files, where literal emoji
 is permitted (and required) as test-input data. See
 [`.github/emoji-prevention.md`](.github/emoji-prevention.md) for the full policy.

## Commit Message Format

```
<type>(<scope>): <subject>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`

**Scopes:** `demojify`, `normalize`, `sanitize`, `docs`, `ci`

**Examples:**
- `feat(sanitize): add RemoveMarkdownFences option`
- `fix(demojify): add missing U+2640-U+2642 gender sign range`
- `docs: update Unicode emoji coverage table`

## Pull Request Checklist

- [ ] `make fmt vet test race` passes
- [ ] Coverage >=80% (`make coverage`)
- [ ] New exported symbols have godoc comments
- [ ] `README.md` updated if API or behavior changed
- [ ] `CHANGELOG.md` entry added under `[Unreleased]`
- [ ] No new external dependencies introduced

## Reporting Issues

Use the [issue templates](.github/ISSUE_TEMPLATE/):

- **Bug report** -- unexpected output or a panic
- **Feature request** -- new sanitization capability
- **Documentation** -- unclear or missing documentation

## License

By contributing, you agree your contributions will be licensed under the same
license as this project. See [LICENSE](LICENSE).