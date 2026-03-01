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

// checkFmt runs gofmt -s -l . and reports unformatted files.
func checkFmt() bool {
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
	fmt.Fprintln(os.Stderr, "[FAIL] gofmt: the following files need formatting (run: make fmt):")
	for _, f := range strings.Split(files, "\n") {
		fmt.Fprintf(os.Stderr, "  %s\n", f)
	}
	return false
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
