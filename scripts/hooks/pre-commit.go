//go:build ignore

// pre-commit.go is the cross-platform implementation of the pre-commit Git
// hook. It is invoked by scripts/hooks/pre-commit via `go run` so the logic
// runs natively on Linux, macOS, and Windows without relying on POSIX
// shell utilities beyond the minimal shebang wrapper.
//
// Install once with:
//
//	make hooks
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if !checkFmt() || !checkVet() || !checkLint() {
		os.Exit(1)
	}
}

// checkFmt runs gofmt -s -l . to find unformatted files, then auto-fixes
// only those files with gofmt -s -w and re-stages them with git add.
// This mirrors what `make fmt` does, so the commit proceeds with clean formatting.
func checkFmt() bool {
	// 1. List files that need formatting.
	out, err := exec.Command("gofmt", "-s", "-l", ".").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] gofmt: %v\n", err)
		return false
	}
	files := strings.TrimSpace(string(out))
	if files == "" {
		fmt.Fprintln(os.Stderr, "[PASS] gofmt")
		return true
	}

	// 2. Collect the specific files that need formatting.
	var unformatted []string
	for _, f := range strings.Split(files, "\n") {
		f = strings.TrimSpace(f)
		if f != "" {
			unformatted = append(unformatted, f)
		}
	}

	// 3. Auto-fix formatting only on the files that need it.
	args := append([]string{"-s", "-w"}, unformatted...)
	fix := exec.Command("gofmt", args...)
	fix.Stderr = os.Stderr
	if err := fix.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] gofmt -w: %v\n", err)
		return false
	}

	// 4. Re-stage the files that were reformatted.
	for _, f := range unformatted {
		add := exec.Command("git", "add", f)
		if err := add.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not re-stage %s: %v\n", f, err)
		}
	}

	fmt.Fprintln(os.Stderr, "[AUTO] gofmt: reformatted and re-staged the following files:")
	for _, f := range unformatted {
		fmt.Fprintf(os.Stderr, "  %s\n", f)
	}
	return true
}

// checkVet runs go vet ./... and reports any issues.
func checkVet() bool {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] go vet (run: make vet)")
		return false
	}
	fmt.Fprintln(os.Stderr, "[PASS] go vet")
	return true
}

// checkLint runs golangci-lint run ./... using the project's .golangci.yml
// configuration. If the binary is not installed the check is skipped with a
// warning so the hook stays portable for contributors who have not installed it.
// The hook also verifies that .golangci.yml exists so that lint always runs
// with the project's intended configuration rather than tool defaults.
func checkLint() bool {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		fmt.Fprintln(os.Stderr, "WARNING: golangci-lint not found -- skipping lint (install: https://golangci-lint.run/usage/install/)")
		return true
	}
	if _, err := os.Stat(".golangci.yml"); err != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] .golangci.yml not found -- create the config file before committing")
		return false
	}
	cmd := exec.Command("golangci-lint", "run", "./...")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] golangci-lint (run: make lint)")
		return false
	}
	fmt.Fprintln(os.Stderr, "[PASS] golangci-lint")
	return true
}
