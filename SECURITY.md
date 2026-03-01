# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest (`main`) | Yes |
| older releases | No |

## Scope

`demojify-sanitize` is a text-processing library with no network access, no file
I/O, no external dependencies, and no execution of user-supplied code. The attack
surface is intentionally minimal:

- **Regex denial-of-service (ReDoS):** All regexes are pre-compiled at package
  init using `regexp.MustCompile`. Go's `regexp` package uses RE2 semantics
  (linear-time matching), which prevents catastrophic backtracking by design.
- **Memory exhaustion:** Passing extremely large strings causes memory use
  proportional to input size. Callers are responsible for applying input size
  limits appropriate to their environment before calling library functions.

## Reporting a Vulnerability

Please **do not open a public GitHub issue** for security vulnerabilities.

Report privately via
[GitHub Security Advisories](https://github.com/nicholashoule/demojify-sanitize/security/advisories/new)
or email the maintainer directly (see GitHub profile).

Include:
- Description of the vulnerability
- Minimal Go code sample to reproduce the issue
- Potential impact assessment

You can expect an initial response within 5 business days.

## Security Best Practices for Consumers

- Apply input size limits before calling library functions if processing
  untrusted input at scale.
- Keep your `go.mod` up to date to receive any future patches.
- Run `govulncheck ./...` in your own CI pipeline to detect transitive
  vulnerabilities in your dependency tree.
