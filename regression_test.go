// Regression tests for the v1.2.0 fixes: the AllowedEmojis placeholder
// collision, writes over a file changed after it was read, and the bounded
// read behind ScanConfig.MaxFileBytes.

package demojify

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// noncharacters are the 34 placeholder sentinels the pre-v1.2.0
// AllowedEmojis implementation tried in turn.
func noncharacters() []rune {
	rs := make([]rune, 0, 34)
	for r := rune(0xFDD0); r <= 0xFDEF; r++ {
		rs = append(rs, r)
	}
	return append(rs, 0xFFFE, 0xFFFF)
}

// TestAllowedEmojisNoPlaceholderCollision is the regression for the
// placeholder swap AllowedEmojis used: hostile input holding every
// noncharacter plus the fallback placeholder had that text "restored" as an
// allowed emoji, and an allowed string matching inside a placeholder (a
// digit) broke the restore. Each case must come back exactly as the input
// minus the emoji that are not allowed.
func TestAllowedEmojisNoPlaceholderCollision(t *testing.T) {
	rocket, grin := "\U0001F680", "\U0001F600"
	var hostile strings.Builder
	for _, r := range noncharacters() {
		hostile.WriteRune(r)
	}
	fallback := "\uFDD0\uFDD1\uFDD2" + "0" + "\uFDD0\uFDD1\uFDD2"
	tests := []struct {
		name    string
		in      string
		allowed []string
		want    string
	}{
		{"every sentinel and the fallback placeholder",
			hostile.String() + fallback + rocket + " launch " + grin, []string{rocket},
			hostile.String() + fallback + rocket + " launch "},
		{"fallback placeholder alone",
			fallback + " " + rocket + grin, []string{rocket},
			fallback + " " + rocket},
		{"allowed digit inside a placeholder",
			rocket + " x1 " + grin, []string{rocket + rocket, rocket, "1"},
			rocket + " x1 "},
		{"allowed sentinel rune itself",
			"\uFDD0" + rocket + grin, []string{rocket, "\uFDD0"},
			"\uFDD0" + rocket},
	}
	for _, tt := range tests {
		for _, aware := range []bool{false, true} {
			opts := Options{RemoveEmojis: true, ScriptAware: aware, AllowedEmojis: tt.allowed}
			if got := Sanitize(tt.in, opts); got != tt.want {
				t.Errorf("%s (ScriptAware=%v): Sanitize = %q, want %q", tt.name, aware, got, tt.want)
			}
		}
	}
}

func TestReadCapped(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		limit    int64
		wantOver bool
	}{
		{"empty", 0, 4, false},
		{"under", 3, 4, false},
		{"exactly at the limit", 4, 4, false},
		{"one over", 5, 4, true},
		{"far over", 1 << 16, 4, true},
		{"zero limit, empty", 0, 0, false},
		{"zero limit, one byte", 1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.Repeat("x", tt.size)
			got, over, err := readCapped(strings.NewReader(in), tt.limit)
			if err != nil {
				t.Fatal(err)
			}
			if over != tt.wantOver {
				t.Fatalf("over = %v, want %v", over, tt.wantOver)
			}
			if !over && string(got) != in {
				t.Errorf("data = %q, want %q", got, in)
			}
			if over && got != nil {
				t.Errorf("data = %d bytes on an over-limit read, want nil", len(got))
			}
		})
	}
}

// TestStatAndWriteRefusesChangedFile is the regression for writes that
// clobbered a file another writer changed after it was read: every
// write-back path (WriteFinding, FixDir, SanitizeFile, ReplaceFile) goes
// through statAndWrite, which must refuse and leave the file alone.
func TestStatAndWriteRefusesChangedFile(t *testing.T) {
	tests := []struct {
		name     string
		onDisk   string
		original string
		wantErr  bool
	}{
		{"unchanged", "a \U0001F600\n", "a \U0001F600\n", false},
		{"edited", "a \U0001F600 edited\n", "a \U0001F600\n", true},
		{"truncated", "a", "a \U0001F600\n", true},
		{"grew", "a \U0001F600\nmore\n", "a \U0001F600\n", true},
		{"emptied", "", "a \U0001F600\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "f.md")
			if err := os.WriteFile(path, []byte(tt.onDisk), 0o640); err != nil {
				t.Fatal(err)
			}
			err := statAndWrite(path, tt.original, "a \n")
			got, _ := os.ReadFile(path)
			if tt.wantErr {
				if !errors.Is(err, ErrFileChanged) {
					t.Fatalf("err = %v, want ErrFileChanged", err)
				}
				if string(got) != tt.onDisk {
					t.Errorf("file = %q, want the other writer's %q untouched", got, tt.onDisk)
				}
				matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".demojify-*"))
				if len(matches) != 0 {
					t.Errorf("temp files left behind: %v", matches)
				}
				return
			}
			if err != nil || string(got) != "a \n" {
				t.Errorf("unchanged file: err = %v, content %q", err, got)
			}
		})
	}
}

// TestWriteFindingRefusesStaleFinding drives the public path: a file edited
// between ScanFile and WriteFinding keeps the edit.
func TestWriteFindingRefusesStaleFinding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(path, []byte("ship it \u2705\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := ScanFile(path, Options{RemoveEmojis: true})
	if err != nil || f == nil {
		t.Fatalf("ScanFile = %v, %v", f, err)
	}
	edited := "ship it \u2705\nnew line from another writer\n"
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := WriteFinding(path, *f)
	if changed || !errors.Is(err, ErrFileChanged) {
		t.Fatalf("WriteFinding = %v, %v; want false, ErrFileChanged", changed, err)
	}
	if got, _ := os.ReadFile(path); string(got) != edited {
		t.Errorf("file = %q, want the edit kept", got)
	}

	// A fresh scan of the edited file writes normally.
	f, err = ScanFile(path, Options{RemoveEmojis: true})
	if err != nil || f == nil {
		t.Fatalf("ScanFile = %v, %v", f, err)
	}
	if changed, err := WriteFinding(path, *f); !changed || err != nil {
		t.Fatalf("WriteFinding(fresh) = %v, %v", changed, err)
	}
}

// TestScanDirCapsSymlinkedLargeFile is the regression for the read behind
// MaxFileBytes being unbounded: the size check reads the walk entry's own
// (Lstat) size, so a symlink to a large file passed it and the file was read
// whole. The read is now bounded, so the file is skipped.
func TestScanDirCapsSymlinkedLargeFile(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	big := filepath.Join(outside, "big.md")
	if err := os.WriteFile(big, []byte(strings.Repeat("\u2705 x\n", 4096)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(big, filepath.Join(root, "link.md")); err != nil {
		t.Skipf("cannot create a symlink here: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "small.md"), []byte("\u2705\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultScanConfig()
	cfg.Root = root
	cfg.MaxFileBytes = 1024
	findings, err := ScanDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Path != "small.md" {
		t.Errorf("findings = %+v, want only small.md (the symlinked file is over the limit)", findings)
	}
	// With the limit disabled (the stable v1 meaning of zero), it is read.
	cfg.MaxFileBytes = 0
	if findings, err = ScanDir(cfg); err != nil || len(findings) != 2 {
		t.Errorf("unlimited scan = %d findings, %v; want 2", len(findings), err)
	}
}

// errReader fails every read.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

// TestWriteHelpersFailClosed covers the error paths: a failing read, a
// target that disappeared, and an atomic write whose target is gone leave
// nothing behind and write nothing.
func TestWriteHelpersFailClosed(t *testing.T) {
	if _, _, err := readCapped(errReader{}, 8); err == nil {
		t.Error("readCapped(failing reader) = nil error")
	}
	dir := t.TempDir()
	missing := filepath.Join(dir, "gone.md")
	if same, err := unchanged(missing, "x"); same || err == nil {
		t.Errorf("unchanged(missing) = %v, %v; want false and an error", same, err)
	}
	if err := atomicWrite(missing, "clean", 0o644, "x"); err == nil {
		t.Error("atomicWrite(missing target) = nil, want an error")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("atomicWrite left %d entries behind, want none", len(entries))
	}
	if err := statAndWrite(missing, "x", "clean"); err == nil {
		t.Error("statAndWrite(missing) = nil, want an error")
	}
}
