//go:build ignore

// installhooks.go installs scripts/hooks/pre-commit into the repository's
// Git hooks directory and marks it executable. It is the cross-platform
// implementation of `make hooks`, replacing a `cp` + `chmod` shell recipe
// that fails on a cmd.exe-based make on Windows (cmd.exe has no cp/chmod).
//
// The hooks directory is resolved through Git rather than hard-coded to
// .git/hooks, so it is correct for linked worktrees and a .git file that
// points at a separate gitdir, and honors an explicit core.hooksPath. If
// the command is not run inside a Git work tree it fails loudly instead of
// fabricating a stray .git/ directory.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[FAIL] "+format+"\n", args...)
	os.Exit(1)
}

// resolveHooksDir returns the directory Git uses for hooks in the current
// repository. It requires being inside a work tree, honors an explicit
// core.hooksPath, and otherwise asks Git for the real hooks path (which
// correctly follows a .git file / linked worktree).
func resolveHooksDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return "", fmt.Errorf("not inside a Git work tree -- run from the repository root")
	}
	// An explicit core.hooksPath overrides the default location. `git
	// config --get` exits non-zero when the key is unset, so only use the
	// value when the command succeeds and is non-empty.
	if out, err := exec.Command("git", "config", "--get", "core.hooksPath").Output(); err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			return filepath.FromSlash(p), nil
		}
	}
	// Resolve the actual hooks dir; this follows a .git file and linked
	// worktrees instead of assuming a literal .git/hooks directory.
	out, err = exec.Command("git", "rev-parse", "--git-path", "hooks").Output()
	if err != nil {
		return "", fmt.Errorf("resolve hooks path via git: %w", err)
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", fmt.Errorf("git returned an empty hooks path")
	}
	return filepath.FromSlash(p), nil
}

func main() {
	src := filepath.Join("scripts", "hooks", "pre-commit")
	data, err := os.ReadFile(src)
	if err != nil {
		fail("read %s: %v", src, err)
	}

	hooksDir, err := resolveHooksDir()
	if err != nil {
		fail("%v", err)
	}

	// The hooks dir normally exists, but a custom core.hooksPath may not;
	// create it without fabricating a parent .git/ (the path is now the
	// Git-resolved location, not a hard-coded .git/hooks).
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		fail("create %s: %v", hooksDir, err)
	}
	dst := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		fail("write %s: %v", dst, err)
	}
	// os.WriteFile only applies perm when creating the file, so chmod
	// explicitly to also fix an existing non-executable hook. On Windows
	// this only toggles the read-only bit (0o755 is writable, so it is a
	// no-op there); Git runs the hook via its bundled sh.exe regardless.
	if err := os.Chmod(dst, 0o755); err != nil {
		fail("chmod %s: %v", dst, err)
	}
	fmt.Fprintf(os.Stderr, "[PASS] pre-commit hook installed at %s\n", dst)
}
