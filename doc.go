// Package demojify helps developers of web applications and APIs audit,
// detect, and fix emoji clutter, AI-generated preamble phrases, and
// redundant whitespace before content reaches production. It is designed
// for two primary workflows: AI agents can import it to self-correct
// their own output, and applications can run it as a gate in their
// request or CI pipeline to catch issues in one pass.
//
// Three primary entry points are exposed:
//
//   - [Demojify] strips emoji and Unicode pictographic characters.
//   - [ContainsEmoji] detects whether text contains emoji.
//   - [Sanitize] applies a configurable pipeline controlled by [Options].
//   - [Normalize] collapses redundant whitespace and blank lines.
//
// For the most common use-case, pass [DefaultOptions] to [Sanitize]:
//
//	clean := demojify.Sanitize(text, demojify.DefaultOptions())
package demojify
