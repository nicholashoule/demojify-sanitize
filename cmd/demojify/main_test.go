package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testBinary is the path to the compiled CLI binary used by all tests.
var testBinary string

func TestMain(m *testing.M) {
	// Build the CLI binary once for all integration tests.
	tmp, err := os.MkdirTemp("", "demojify-cli-test-*")
	if err != nil {
		panic(err)
	}
	bin := filepath.Join(tmp, "demojify")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build CLI binary: " + err.Error())
	}
	testBinary = bin

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// writeTempFile creates a file inside dir with the given name and content.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// runCLI executes the test binary with the given args and returns stdout,
// stderr, and the exit code. A non-zero exit code is not treated as an error.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(testBinary, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("exec error: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}
