# Contributing to demojify-sanitize

Thank you for your interest in contributing. This is a focused, dependency-free
Go library for text sanitization. Contributions should remain within that scope.

## Before You Start

- Search [existing issues](https://github.com/nicholashoule/demojify-sanitize/issues).
- For non-trivial changes, open an issue first to discuss the approach.
- Keep changes within the library's dependency-free text-sanitization scope.

## Development Setup

```bash
git clone https://github.com/nicholashoule/demojify-sanitize.git
cd demojify-sanitize
make hooks    # install the repository pre-commit hook once
make test     # confirm the baseline passes
```

Development requires Go 1.21 or newer. `make hooks` installs the tracked hook,
which runs repository governance, emoji auditing, formatting, vet, lint when
available, and tests. Run the same Go quality checks without installing it via
`make pre-commit`. The installed hook is a copy, so re-run `make hooks`
whenever `scripts/hooks/pre-commit` changes.

Optional tools:

- `make` for the documented convenience targets.
- `golangci-lint` for `make lint` ([installation](https://golangci-lint.run/usage/install/)).

`make race` requires CGO and a C compiler. On Windows, install GCC through
[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/).
CI runs race-enabled tests across its supported platform matrix.

## Making Changes

1. Fork the repository and create a feature branch.
2. Install the hook with `make hooks`.
3. Write or update focused tests.
4. Update `README.md` and `CHANGELOG.md` if the public API or behavior changes.
5. Run `make fmt vet test race` and, when installed, `make lint`.
6. Open a pull request against `main`.

## Code Standards

- **No external dependencies.** This is a hard requirement.
- **Compiled regexes at package level.** All `regexp.MustCompile` calls must be
 `var` declarations, never inside functions.
- **Table-driven tests.** See existing `*_test.go` files for the pattern.
- **`gofmt -s` formatting.** Run `make fmt` before committing.
- **Godoc comments on all exported symbols.**
- **No emoji in production source, comments, or output.** Use `[PASS]`, `[FAIL]`,
  `WARNING:`, and similar text. Literal emoji are permitted in `*_test.go`
  fixtures. See [the emoji policy](.github/rules/emoji-prevention.md).

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
- [ ] `make lint` passes when `golangci-lint` is available
- [ ] New exported symbols have godoc comments
- [ ] Public API or behavior changes are documented
- [ ] User-visible changes are reflected in `CHANGELOG.md`
- [ ] No new external dependencies introduced

## Reporting Issues

Open a [GitHub issue](https://github.com/nicholashoule/demojify-sanitize/issues)
with reproduction steps for bugs or a clear use case for feature requests.

## License

By contributing, you agree your contributions will be licensed under the same
license as this project. See [LICENSE](LICENSE).