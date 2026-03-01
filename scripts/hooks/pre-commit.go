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
	out, err := exec.Command("gofmt", "-s", "-l", ".").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] gofmt: %v\n", err)
		os.Exit(1)
	}
	files := strings.TrimSpace(string(out))
	if files == "" {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, "[FAIL] gofmt: the following files need formatting (run: make fmt):")
	for _, f := range strings.Split(files, "\n") {
		fmt.Fprintf(os.Stderr, "  %s\n", f)
	}
	os.Exit(1)
}
