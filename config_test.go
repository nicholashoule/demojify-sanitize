package demojify_test

import (
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func TestDefaultLimitConfig(t *testing.T) {
	cfg := demojify.DefaultLimitConfig()
	if cfg.Default != 500 {
		t.Errorf("DefaultLimitConfig().Default = %d, want 500", cfg.Default)
	}
	if v, ok := cfg.Files[".claude/CLAUDE.md"]; !ok || v != 50 {
		t.Errorf("DefaultLimitConfig().Files[\".claude/CLAUDE.md\"] = %d (ok=%v), want 50 (ok=true)", v, ok)
	}
}

func TestDefaultLineLimit(t *testing.T) {
	if demojify.DefaultLineLimit != 500 {
		t.Errorf("DefaultLineLimit = %d, want 500", demojify.DefaultLineLimit)
	}
}

func TestLimitConfig_ZeroDefaultFallback(t *testing.T) {
	// A zero Default should fall back to DefaultLineLimit via DefaultLimitConfig.
	cfg := demojify.DefaultLimitConfig()
	if cfg.Default != demojify.DefaultLineLimit {
		t.Errorf("DefaultLimitConfig().Default = %d, want DefaultLineLimit (%d)", cfg.Default, demojify.DefaultLineLimit)
	}
}

func TestLimitConfig_FileOverride(t *testing.T) {
	cfg := demojify.LimitConfig{
		Default: 200,
		Files:   map[string]int{".claude/CLAUDE.md": 50},
	}
	if got := cfg.Files[".claude/CLAUDE.md"]; got != 50 {
		t.Errorf("Files[\".claude/CLAUDE.md\"] = %d, want 50", got)
	}
	if cfg.Default != 200 {
		t.Errorf("Default = %d, want 200", cfg.Default)
	}
}

func TestResolveLimit(t *testing.T) {
	tests := []struct {
		name string
		cfg  demojify.LimitConfig
		path string
		want int
	}{
		{
			name: "file override takes precedence",
			cfg:  demojify.LimitConfig{Default: 200, Files: map[string]int{".claude/CLAUDE.md": 50}},
			path: ".claude/CLAUDE.md",
			want: 50,
		},
		{
			name: "default used when no file override",
			cfg:  demojify.LimitConfig{Default: 300},
			path: "any/file.md",
			want: 300,
		},
		{
			name: "zero default falls back to DefaultLineLimit",
			cfg:  demojify.LimitConfig{Default: 0},
			path: "any/file.go",
			want: demojify.DefaultLineLimit,
		},
		{
			name: "backslash path normalised to forward slash",
			cfg:  demojify.LimitConfig{Default: 100, Files: map[string]int{"a/b.go": 42}},
			path: `a\b.go`,
			want: 42,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := demojify.ResolveLimit(tt.cfg, tt.path); got != tt.want {
				t.Errorf("ResolveLimit(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}
