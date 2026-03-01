---
applyTo: "**"
---

# Code Review Instructions

Perform a full, end-to-end review of this Go module as a Sr. developer or SRE.
Treat it as production-grade: idiomatic, stable, dependency-free, and ready for
downstream consumers. The review output serves as an auditable proof of quality.

**Output location:** `docs/review.md`

**Related instructions:**
[general](general.instructions.md) |
[library](library.instructions.md) |
[testing](testing.instructions.md)

**Project constraints** (enforced during review):
- Zero external dependencies -- stdlib only
- No emoji in production code (test-input data in `*_test.go` is exempt)
- Compiled regexes at package-level `var`, never inside functions
- Pure text APIs (`Demojify`, `Normalize`, `Sanitize`) never return errors
- Backward compatibility -- do not remove or rename exported symbols

## Severity Classification

| Level | Meaning |
|-------|---------|
| Critical | Build failure, data loss, security issue, or API breakage |
| High | Incorrect behavior, race condition, or missing error handling |
| Low | Style, naming, minor inconsistency, or missing edge-case test |
| Info | Observation or suggestion; no action required |

## Review Tasks

### 1. Module / API Review

Evaluate every exported symbol against these criteria:

- **Naming:** idiomatic Go, consistent with stdlib conventions
- **Signatures:** correct types, no unnecessary `interface{}`
- **Error handling:** I/O functions return `error`; pure text functions do not
- **Documentation:** every exported symbol has a godoc comment
- **Concurrency safety:** defaults return independent copies
- **Internal structure:** single-responsibility files, no circular logic

Report strengths as a bullet list. Report findings in a table with columns:
`ID | Severity | File | Description`. Use IDs `A1`, `A2`, etc.

### 2. Test Review

Assess against [testing.instructions.md](testing.instructions.md):

- **Structure:** table-driven, external test package (`demojify_test`)
- **Coverage:** >= 80% target; report exact percentage
- **Edge cases:** empty input, nil maps, binary files, permission preservation,
  idempotency, ZWJ sequences, variation selectors, non-emoji Unicode
- **Dogfooding:** `repo_test.go` scans the repo with its own API
- **Examples:** `example_test.go` covers public API for pkg.go.dev

Report findings in a table with IDs `T1`, `T2`, etc. For each missing test,
state what should be tested and mark it **Added** once implemented.

### 3. Real-World Validation

Run each check and report PASS or FAIL in a table:

| Check | Command |
|-------|---------|
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Tests | `go test ./...` |
| Race detector | `go test ./... -race` |
| Coverage | `go test ./... -cover` |
| Formatting | `gofmt -l .` |
| CLI build | `go build -o demojify ./cmd/demojify/` |
| Driver program | `go run ./docs/examples/driver/` |
| Self-validation | `demojify -root docs -exts .md` |

All checks must pass for the review to be considered clean.
The self-validation step runs the module's own CLI against every output file
produced by the review (e.g., `docs/review.md`). Any emoji found must be
substituted (`-sub`) or stripped (`-fix`) before the review is final.

### 4. Driver Program

Create or verify `docs/examples/driver/main.go`. It must:

- Import the module as a downstream consumer would
- Exercise every major public API (detect, strip, normalize, sanitize,
  replace, count, find, scan, write-back)
- Create temp files for I/O operations
- Verify idempotency (re-scan after write-back finds zero issues)
- Compile and exit 0

List the APIs demonstrated as a numbered list.

### 5. CLI Validation

Test every flag combination from `cmd/demojify/main.go`:

| Flag | Behavior to verify |
|------|--------------------|
| `-root <dir>` | Scans directory; exit 0 if clean, 1 if findings |
| `-fix` | Rewrites files in place; exit 0 |
| `-sub` | Substitutes emoji with text tokens; implies `-fix` |
| `-normalize` | Collapses whitespace; implies `-fix` |
| `-quiet` | No stdout; exit code only |
| `-exts .go,.md` | Filters by extension; auto-prepends dot if missing |
| `-root /nonexistent` | Fatal error with descriptive message |

Report as a table: `Command | Expected | Actual | Status`.
Note any documentation inconsistencies (e.g., redundant flags in examples).

### 6. CLI Test Examples

Provide runnable CLI invocations with expected output. Cover:
- Audit only (no writes)
- Strip emoji (`-fix`)
- Substitute (`-sub`)
- Quiet mode (`-quiet`)
- Extension filter (`-exts`)
- Error case (bad root)

Validate each example runs against the current codebase.

### 7. Recommendations

Summarize as a numbered list. Classify each as:
- **Action required** -- must fix before release
- **Recommended** -- should fix; improves quality
- **Deferred** -- tracked for a future release

Mark resolved items with strikethrough and the version where fixed.

## Output Format

The review document (`docs/review.md`) must follow this structure:

```
# Module Review: demojify-sanitize
## Summary          -- one-paragraph verdict with key metrics
## 1. Module / API Review
  ### Strengths     -- bullet list
  ### Findings      -- ID | Severity | File | Description table
## 2. Test Review
  ### Strengths     -- bullet list
  ### Findings      -- ID | Severity | File | Description table
  ### Coverage      -- exact percentage and uncovered paths
## 3. Real-World Validation  -- Check | Result table
## 4. Driver Program         -- numbered API list
## 5. CLI Validation
  ### Flag behavior          -- Command | Expected | Actual | Status table
  ### Help text              -- assessment
  ### Observations           -- bullet list
## 6. CLI Test Examples      -- runnable code blocks
## 7. Recommendations        -- numbered, classified list
```

## Acceptance Criteria

The review is complete when:
- [ ] All validation checks in section 3 report PASS
- [ ] Every exported symbol is assessed in section 1
- [ ] Test coverage is reported with exact percentage (>= 80%)
- [ ] Driver program compiles and exits 0
- [ ] Every CLI flag is tested with expected vs. actual behavior
- [ ] Findings use consistent IDs and severity levels
- [ ] Recommendations are classified and actionable
- [ ] `docs/review.md` follows the output format above
- [ ] Self-validation passes: `demojify -root docs -exts .md` reports no emoji
