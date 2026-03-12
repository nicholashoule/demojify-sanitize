package demojify

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Default != 500 {
		t.Errorf("DefaultConfig().Default = %d, want 500", cfg.Default)
	}
	if v, ok := cfg.Files[".claude/CLAUDE.md"]; !ok || v != 50 {
		t.Errorf("DefaultConfig().Files[\".claude/CLAUDE.md\"] = %d (ok=%v), want 50 (ok=true)", v, ok)
	}
}

func TestResolveLimit_ZeroDefault(t *testing.T) {
	cfg := LimitConfig{Default: 0}
	got := resolveLimit(cfg, "any/file.go")
	if got != DefaultLineLimit {
		t.Errorf("resolveLimit(cfg with zero default) = %d, want %d (DefaultLineLimit)", got, DefaultLineLimit)
	}
}

func TestResolveLimit_FileOverride(t *testing.T) {
	cfg := LimitConfig{
		Default: 200,
		Files:   map[string]int{".claude/CLAUDE.md": 50},
	}
	if got := resolveLimit(cfg, ".claude/CLAUDE.md"); got != 50 {
		t.Errorf("resolveLimit for .claude/CLAUDE.md = %d, want 50", got)
	}
	if got := resolveLimit(cfg, "other.go"); got != 200 {
		t.Errorf("resolveLimit for other.go = %d, want 200", got)
	}
}

func TestResolveLimit_DefaultUsed(t *testing.T) {
	cfg := LimitConfig{Default: 300}
	if got := resolveLimit(cfg, "any/file.md"); got != 300 {
		t.Errorf("resolveLimit = %d, want 300", got)
	}
}
