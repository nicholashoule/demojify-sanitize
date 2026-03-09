# Git Hooks Integration

Integrate `demojify-sanitize` into a Git pre-commit hook to catch emoji before
code is committed. No third-party frameworks required.

## CLI hook (recommended)

The simplest approach -- no Go code needed. First build a local binary from
this repository so the hook never fetches remote code:

```sh
go build -o .git/hooks/demojify ./cmd/demojify
```

Or install a pinned release (replace `vX.Y.Z` with a specific tag):

```sh
go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@vX.Y.Z
# then copy or symlink the installed binary into .git/hooks/
```

To also enforce line-length and layout governance, pair with
[repogov](https://github.com/nicholashoule/repogov) (see the cross-platform
examples at the bottom of this file). The hook skips repogov gracefully when
the sibling directory is absent, so it is safe to add unconditionally.

Audit-only hook (blocks the commit if emoji are found):

```sh
#!/bin/sh
# .git/hooks/pre-commit
root="$(git rev-parse --show-toplevel)"
"$root/.git/hooks/demojify" \
  -root "$root" \
  -exts .go,.md \
  -quiet
```

Exit 1 blocks the commit; exit 0 allows it.

Auto-fix instead of blocking (rewrite working tree, then re-stage):

```sh
#!/bin/sh
# .git/hooks/pre-commit
root="$(git rev-parse --show-toplevel)"
"$root/.git/hooks/demojify" \
  -root "$root" \
  -exts .go,.md \
  -quiet || {
    echo "ERROR: emoji found -- fixing and re-staging:"
    "$root/.git/hooks/demojify" -root "$root" -exts .go,.md -sub
    git add -u
    exit 1
  }
```

NOTE: The hook exits 1 after fixing so the developer can review the rewrite
before the commit is retried. Remove `exit 1` to allow the commit to proceed
automatically after sanitization.

Make it executable after saving:

```sh
chmod +x .git/hooks/pre-commit
```

## Go API hook

For programmatic control (custom logic, extension filters, chaining checks),
use `ScanDir` or `FixDir` directly in a `//go:build ignore` Go file invoked
from the shell shim:

```sh
#!/bin/sh
# .git/hooks/pre-commit
exec go run "$(git rev-parse --show-toplevel)/scripts/hooks/pre-commit.go"
```

See [scripts/hooks/pre-commit.go](../scripts/hooks/pre-commit.go) for this
repository's working example and [docs/examples/driver/main.go](examples/driver/main.go)
for full API usage patterns.

---

## Cross-platform lightweight examples

These minimal hooks mirror the pattern used in `scripts/hooks/pre-commit` and
run both tools from their published module versions -- no local clone required.

### macOS and Linux (`sh`)

```sh
#!/bin/sh
# .git/hooks/pre-commit
root="$(git rev-parse --show-toplevel)"
cd "$root"

go run github.com/nicholashoule/repogov/cmd/repogov@v0.2.0 -root "$root" -agent copilot
repogov_exit=$?

go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.4.0 -root "$root"
demojify_exit=$?

exit $((repogov_exit | demojify_exit))
```

### Windows (PowerShell -- save as `pre-commit.ps1`, invoked by the shim below)

```powershell
# scripts/hooks/pre-commit.ps1
$root = git rev-parse --show-toplevel
Set-Location $root

go run github.com/nicholashoule/repogov/cmd/repogov@v0.2.0 -root $root -agent copilot
$repogov_exit = $LASTEXITCODE

go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@v0.4.0 -root $root
$demojify_exit = $LASTEXITCODE

exit ($repogov_exit -bor $demojify_exit)
```

Git on Windows requires the hook file itself to be a POSIX-style `sh` script.
Use this one-line shim as `.git/hooks/pre-commit` to delegate to the
PowerShell script:

```sh
#!/bin/sh
powershell -NoProfile -ExecutionPolicy Bypass -File \
  "$(git rev-parse --show-toplevel)/scripts/hooks/pre-commit.ps1"
```

### Windows (Git for Windows `sh.exe` -- no shim needed)

Git for Windows ships `sh.exe`; the macOS/Linux hook above works unchanged.
Install it the same way:

```sh
cp scripts/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

