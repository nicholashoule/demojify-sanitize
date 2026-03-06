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
	if !checkFmt() || !checkVet() {
		os.Exit(1)
	}
}

// checkFmt runs gofmt -s -l . to find unformatted files, then auto-fixes them
// with gofmt -s -w . and re-stages the changed files with git add.
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

	// 2. Auto-fix formatting in place.
	fix := exec.Command("gofmt", "-s", "-w", ".")
	fix.Stderr = os.Stderr
	if fix.Run() != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] gofmt -w failed")
		return false
	}

	// 3. Re-stage the files that were reformatted.
	changed := strings.Split(files, "\n")
	for _, f := range changed {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if add := exec.Command("git", "add", f); add.Run() != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not re-stage %s\n", f)
		}
	}

	fmt.Fprintln(os.Stderr, "[AUTO] gofmt: reformatted and re-staged the following files:")
	for _, f := range changed {
		if f = strings.TrimSpace(f); f != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
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
