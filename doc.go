// Package demojify detects, removes, or substitutes emoji-related codepoints
// and normalizes redundant whitespace. It is suitable for AI output, submitted
// content, text files, and CI quality gates. The package has no third-party
// dependencies.
//
// # Text processing
//
// [Sanitize] combines configurable emoji removal and whitespace normalization.
// [Demojify], [ContainsEmoji], [CountEmoji], [BytesSaved], and [Normalize]
// expose the individual operations. For the common case:
//
//	clean := demojify.Sanitize(text, demojify.DefaultOptions())
//
// [SanitizeReport] returns a [SanitizeResult] with removal metrics,
// [SanitizeReader] streams an [io.Reader] line by line, and [SanitizeJSON]
// cleans only the string values of a JSON document, returning
// [ErrMultipleJSONValues] for concatenated documents. [Options.AllowedRanges]
// and [Options.AllowedEmojis] preserve specific codepoints, and
// [TechnicalSymbolRanges] supplies a ready-made allow list.
//
// # Emoji substitution
//
// [Replace] substitutes mapped sequences and removes recognized unmapped emoji.
// [DefaultReplacements] provides the built-in text map; [ReplaceCount],
// [FindAll], and [FindAllMapped] provide reporting helpers.
//
//	repl := demojify.DefaultReplacements()
//	clean := demojify.Replace(text, repl)
//
// # File and directory scanning
//
// [ScanFile], [ScanDir], and [ScanDirContext] return a [Finding] for each
// file whose content would change; [FindMatchesInFile] and
// [ScanConfig.CollectMatches] add per-occurrence [Match] detail.
// [SanitizeFile], [ReplaceFile], [WriteFinding], and [FixDir] provide
// permission-preserving file updates. Use [DefaultScanConfig] as a starting
// point and adjust its filters for the repository being scanned.
//
// # CLI
//
// A standalone command-line tool is provided at
// github.com/nicholashoule/demojify-sanitize/cmd/demojify.
// Install v1 with:
//
//	go install github.com/nicholashoule/demojify-sanitize/cmd/demojify@v1.0.0
//
// See the cmd/demojify package documentation for the full CLI reference,
// including modes, flags, exit codes, and JSON output.
package demojify
