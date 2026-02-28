// Package demojify provides functions to detect and remove emojis,
// Unicode pictographic characters, AI-generated clutter, and other
// non-semantic artifacts from Markdown, documentation, and repository
// text files.
//
// Three primary entry points are exposed:
//
//   - [Demojify] strips emoji and Unicode pictographic characters.
//   - [Sanitize] applies a configurable pipeline controlled by [Options].
//   - [Normalize] collapses redundant whitespace and blank lines.
//
// For the most common use-case, pass [DefaultOptions] to [Sanitize]:
//
//	clean := demojify.Sanitize(text, demojify.DefaultOptions())
package demojify
