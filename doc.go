// Package demojify helps developers of web applications and APIs audit,
// detect, and fix emoji clutter and redundant whitespace in text content
// before it reaches production. AI agents can import it to self-correct
// their own output, and applications can run it as a gate in their
// request or CI pipeline to catch issues in one pass.
//
// # Text processing
//
//   - [Demojify] strips emoji and Unicode pictographic characters.
//   - [ContainsEmoji] detects whether text contains emoji.
//   - [CountEmoji] returns the number of emoji codepoint occurrences in text.
//   - [BytesSaved] reports how many bytes emoji removal would save.
//   - [Sanitize] applies a configurable pipeline controlled by [Options].
//   - [SanitizeReport] applies [Sanitize] and returns a [SanitizeResult] with
//     metrics (emoji removed, bytes saved) for observability in agent pipelines.
//   - [SanitizeReader] applies the sanitization pipeline to an [io.Reader]
//     line by line, writing results to an [io.Writer] for streaming scenarios
//     such as LLM token streams and MCP transports.
//   - [SanitizeJSON] sanitizes only string values within a JSON document,
//     leaving keys, numbers, booleans, and null unchanged.
//   - [SanitizeFile] applies [Sanitize] to a file atomically; no write
//     occurs when the file is already clean.
//   - [WriteFinding] writes a [Finding.Cleaned] result back to disk
//     atomically, avoiding the re-read that [SanitizeFile] or [ReplaceFile]
//     would perform after a [ScanDir] pass.
//   - [Normalize] collapses redundant whitespace and blank lines.
//   - [TechnicalSymbolRanges] returns Unicode range tables for technical
//     symbols (check marks, warning signs, gears, card suits) that fall
//     within the emoji regex but are not clutter; pass to
//     [Options.AllowedRanges] to preserve them.
//
// For the most common use-case, pass [DefaultOptions] to [Sanitize]:
//
//	clean := demojify.Sanitize(text, demojify.DefaultOptions())
//
// # Emoji substitution
//
// Rather than stripping emoji, callers can substitute them with meaningful
// text equivalents using the replacement functions:
//
//   - [Replace] substitutes codepoints using a caller-supplied map, then
//     strips any residual emoji via [Demojify]. Longer keys match first.
//   - [ReplaceFile] applies [Replace] to a file atomically; no write
//     occurs when the file is already clean.
//   - [ReplaceCount] applies [Replace] and also returns the substitution count.
//   - [FindAll] returns distinct emoji sequences found in text.
//   - [FindAllMapped] returns only mapped-key sequences, greedy longest-first.
//   - [DefaultReplacements] returns a built-in ~137-entry emoji-to-text map
//     covering status symbols, arrows, shapes, checkboxes, and dingbats.
//
// Typical usage:
//
//	repl := demojify.DefaultReplacements()
//	clean := demojify.Replace(text, repl)
//
//	// or, to replace and count in one call:
//	clean, n := demojify.ReplaceCount(text, repl)
//
// # File and directory scanning
//
//   - [ScanDir] walks a directory tree and returns a [Finding] for every
//     file whose content would change after sanitization.
//   - [ScanDirContext] is like [ScanDir] but accepts a [context.Context]
//     for cancellation support in agent orchestrators and MCP servers.
//   - [FixDir] scans a directory tree and writes back every modified file
//     in one call, combining [ScanDir] with [WriteFinding].
//   - [ScanFile] checks a single file and returns a [Finding] if it needs
//     sanitization, or nil if it is already clean.
//   - [FindMatchesInFile] returns per-occurrence [Match] detail for a file
//     without sanitizing it.
//   - [ScanConfig] configures directory/file exemptions, extension filters,
//     the sanitization [Options], an optional [ScanConfig.Replacements] map
//     (uses [Replace] instead of [Sanitize] when set), and
//     [ScanConfig.CollectMatches] (populates [Finding.Matches] per file).
//   - [DefaultScanConfig] returns a config that scans all file types with
//     sensible directory and suffix exemptions.
//
// # CLI
//
// A standalone command-line tool is provided at
// github.com/nicholashoule/demojify-sanitize/cmd/demojify.
// Install it with:
//
//	go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest
//
// Or run without installing:
//
//	go run github.com/nicholashoule/demojify-sanitize/cmd/demojify@latest [flags]
//
// See the cmd/demojify package documentation for the full CLI reference,
// including subcommands (operational modes), flags, exit codes, and JSON
// output format.
package demojify
