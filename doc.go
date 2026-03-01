// Package demojify helps developers of web applications and APIs audit,
// detect, and fix emoji clutter, AI-generated preamble phrases, and
// redundant whitespace before content reaches production. It is designed
// for two primary workflows: AI agents can import it to self-correct
// their own output, and applications can run it as a gate in their
// request or CI pipeline to catch issues in one pass.
//
// # Text processing
//
//   - [Demojify] strips emoji and Unicode pictographic characters.
//   - [ContainsEmoji] detects whether text contains emoji.
//   - [Sanitize] applies a configurable pipeline controlled by [Options].
//   - [Normalize] collapses redundant whitespace and blank lines.
//
// For the most common use-case, pass [DefaultOptions] to [Sanitize]:
//
//	clean := demojify.Sanitize(text, demojify.DefaultOptions())
//
// # File and directory scanning
//
//   - [ScanDir] walks a directory tree and returns a [Finding] for every
//     file whose content would change after sanitization.
//   - [ScanFile] checks a single file and returns a [Finding] if it needs
//     sanitization, or nil if it is already clean.
//   - [ScanConfig] configures directory/file exemptions, extension filters,
//     and the sanitization [Options] applied to each file.
//   - [DefaultScanConfig] returns a config that scans all file types with
//     sensible directory and suffix exemptions.
package demojify
