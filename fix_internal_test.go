package demojify

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsInsideDir(t *testing.T) {
	sep := string(filepath.Separator)

	// Use platform-appropriate absolute paths.
	var root, child, nested, sibling, prefixDir, parent string
	if runtime.GOOS == "windows" {
		root = `C:\project\root`
		child = `C:\project\root\sub`
		nested = `C:\project\root\sub\deep\file.txt`
		sibling = `C:\project\other`
		prefixDir = `C:\project\rootExtra`
		parent = `C:\project`
	} else {
		root = "/project/root"
		child = "/project/root/sub"
		nested = "/project/root/sub/deep/file.txt"
		sibling = "/project/other"
		prefixDir = "/project/rootExtra"
		parent = "/project"
	}

	tests := []struct {
		name   string
		target string
		dir    string
		want   bool
	}{
		{
			name:   "same directory",
			target: root,
			dir:    root,
			want:   true,
		},
		{
			name:   "direct child",
			target: child,
			dir:    root,
			want:   true,
		},
		{
			name:   "nested child",
			target: nested,
			dir:    root,
			want:   true,
		},
		{
			name:   "sibling directory",
			target: sibling,
			dir:    root,
			want:   false,
		},
		{
			name:   "dir name is prefix of target",
			target: prefixDir,
			dir:    root,
			want:   false,
		},
		{
			name:   "parent directory",
			target: parent,
			dir:    root,
			want:   false,
		},
		{
			name:   "dot-dot escapes root",
			target: root + sep + ".." + sep + "other",
			dir:    root,
			want:   false,
		},
		{
			name:   "dot-dot then back in stays inside",
			target: root + sep + "sub" + sep + ".." + sep + "sub",
			dir:    root,
			want:   true,
		},
		{
			name:   "bare dot-dot is parent",
			target: filepath.Dir(root),
			dir:    root,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInsideDir(tt.target, tt.dir)
			if got != tt.want {
				t.Errorf("isInsideDir(%q, %q) = %v, want %v", tt.target, tt.dir, got, tt.want)
			}
		})
	}
}
