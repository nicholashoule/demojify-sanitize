//go:build ignore

// clean.go removes build and coverage artifacts produced by `make
// coverage`. It is the cross-platform implementation of `make clean`,
// replacing `rm -f` which is unavailable on a cmd.exe-based make on
// Windows. A missing file is not an error, matching `rm -f` semantics.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	for _, name := range []string{"coverage.out", "coverage.html"} {
		if err := os.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "[FAIL] remove %s: %v\n", name, err)
			os.Exit(1)
		}
	}
	fmt.Fprintln(os.Stderr, "[PASS] clean")
}
