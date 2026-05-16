//go:build ignore

// fmtcheck.go reports Go files that are not gofmt-simplified, without
// modifying them. It is the cross-platform implementation of `make
// fmt-check`, replacing a POSIX-only shell recipe (out=$(...); [ -n ... ])
// that fails when make runs recipes via cmd.exe on Windows.
//
// Unlike scripts/hooks/pre-commit.go, this only reports; it never runs
// gofmt -w or re-stages files, preserving the original fmt-check contract.
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
		fmt.Fprintln(os.Stderr, "[PASS] gofmt")
		return
	}
	fmt.Fprintln(os.Stderr, "[FAIL] gofmt -- the following files need formatting (run: make fmt):")
	for _, f := range strings.Split(files, "\n") {
		if f = strings.TrimSpace(f); f != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
	}
	os.Exit(1)
}
