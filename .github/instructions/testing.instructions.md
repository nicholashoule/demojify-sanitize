---
applyTo: "**/*_test.go"
---

# Testing Instructions

## Framework

**Go `testing` package** with table-driven tests and runnable `Example*` functions.

## File Layout

| File | Tests |
|------|-------|
| `demojify_test.go` | `TestDemojify`, `TestContainsEmoji` |
| `normalize_test.go` | `TestNormalize` |
| `sanitize_test.go` | `TestDefaultOptions`, `TestSanitize` |
| `example_test.go` | `Example*` functions (rendered on pkg.go.dev) |

All tests are in package `demojify_test` (external test package) to verify the public API only.

## Running Tests

```bash
# All tests
go test ./...

# Verbose output
go test ./... -v

# Race detector (required before PR)
go test ./... -race

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Using Make
make test          # go test ./...
make race          # go test ./... -race
make coverage      # coverage report
```

## Table-Driven Test Pattern

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {
            name:  "descriptive case name",
            input: "input value",
            want:  "expected output",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := demojify.FunctionName(tt.input)
            if got != tt.want {
                t.Errorf("FunctionName(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}
```

## Writing Example Functions

Examples appear on `pkg.go.dev` and are verified by `go test`.
Always include an expected output comment:

```go
func ExampleFunctionName() {
    fmt.Println(demojify.FunctionName("input"))
    // Output:
    // expected output
}
```

## Test Structure Rules

- Use table-driven tests for all functions with more than one meaningful input
- Use `t.Run()` for subtests with human-readable names (not camelCase)
- Test both expected behavior and edge cases (empty string, no match, boundary)
- Test false-positive prevention -- e.g., a phrase NOT at line start is preserved
- Do not test unexported symbols; test only through the public API

## Assertion Style

```go
// String comparison
if got != want {
    t.Errorf("Function(%q) = %q, want %q", input, got, want)
}

// Boolean comparison
if got != want {
    t.Errorf("ContainsEmoji(%q) = %v, want %v", input, got, want)
}
```

Prefer `t.Errorf` (non-fatal) over `t.Fatalf` unless the test cannot continue.

## Coverage Target

**>80%** overall. Check with:

```bash
go test ./... -cover
```

## Quality Gates

Before committing:
- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test ./... -race`
- [ ] Examples verified: `go test ./... -v` (look for `--- PASS: ExampleXxx`)
- [ ] Coverage >=80%: `go test ./... -cover`

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Testable Examples](https://go.dev/blog/examples)
