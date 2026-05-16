//go:build ignore

// installhooks.go copies scripts/hooks/pre-commit into .git/hooks/pre-commit
// and marks it executable. It is the cross-platform implementation of
// `make hooks`, replacing a `cp` + `chmod` shell recipe that fails on a
// cmd.exe-based make on Windows (cmd.exe has no cp or chmod).
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	src := filepath.Join("scripts", "hooks", "pre-commit")
	dstDir := filepath.Join(".git", "hooks")
	dst := filepath.Join(dstDir, "pre-commit")

	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] read %s: %v\n", src, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] create %s: %v\n", dstDir, err)
		os.Exit(1)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] write %s: %v\n", dst, err)
		os.Exit(1)
	}
	// os.WriteFile only applies perm when creating the file, so chmod
	// explicitly to also fix an existing non-executable hook. On Windows
	// this only toggles the read-only bit (0o755 is writable, so it is a
	// no-op there); Git runs the hook via its bundled sh.exe regardless.
	if err := os.Chmod(dst, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] chmod %s: %v\n", dst, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "[PASS] pre-commit hook installed")
}
