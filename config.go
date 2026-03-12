package demojify

import "path/filepath"

// LimitConfig specifies per-file line limits used when scanning files in a
// repository. A zero Default field falls back to DefaultLineLimit.
// Use [DefaultConfig] to obtain a pre-populated config suitable for most
// Go module and AI-agent workspaces.
type LimitConfig struct {
	// Default is the maximum number of lines to scan per file.
	// When zero, [DefaultLineLimit] is used as the fallback.
	Default int

	// Files maps individual file paths (relative, forward-slash formatted)
	// to their per-file line limits, overriding Default for those files.
	Files map[string]int
}

// DefaultLineLimit is the fallback line limit applied when LimitConfig.Default
// is zero.
const DefaultLineLimit = 500

// DefaultConfig returns a LimitConfig with sensible defaults for typical Go
// module and AI-agent workspace scanning:
//   - Default line limit of 500
//   - .claude/CLAUDE.md limited to 50 lines (short AI context file)
func DefaultConfig() LimitConfig {
	return LimitConfig{
		Default: 500,
		Files: map[string]int{
			".claude/CLAUDE.md": 50,
		},
	}
}

// resolveLimit returns the effective line limit for path using cfg.
// A file-specific entry in cfg.Files takes precedence over cfg.Default.
// When cfg.Default is zero the fallback [DefaultLineLimit] is returned.
func resolveLimit(cfg LimitConfig, path string) int {
	if v, ok := cfg.Files[filepath.ToSlash(path)]; ok {
		return v
	}
	if cfg.Default == 0 {
		return DefaultLineLimit
	}
	return cfg.Default
}
