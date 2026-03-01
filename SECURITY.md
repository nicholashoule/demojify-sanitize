# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest (`main`) | Yes |
| older releases | No |

## Scope

`demojify-sanitize` is a text-processing library with no network access, no
external dependencies, and no execution of user-supplied code.

The package has two distinct tiers with different attack surfaces:

**Pure text-processing functions** (`Demojify`, `ContainsEmoji`, `Normalize`,
`Sanitize`): no file I/O. Input is an in-memory string; output is a string.
The attack surface is limited to regex and Unicode processing on caller-supplied
data.

**Scanner functions** (`ScanDir`, `ScanFile`): perform filesystem reads and
directory walking via `os.ReadFile` and `filepath.WalkDir`. The attack surface
includes path traversal (callers should validate `ScanConfig.Root`), large or
binary files (mitigated by `ScanConfig.MaxFileBytes`, defaulting to 1 MiB), and
the same regex/memory considerations as the text-processing tier.

Shared security properties:

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
