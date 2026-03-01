package demojify_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// benchInput builds a string with n emoji interspersed in plain text.
func benchInput(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("build step ")
		b.WriteString("\u2705 passed, deploy \U0001F680 done, check \U0001F4CA report\n")
	}
	return b.String()
}

// buildLargeMarkdown generates a realistic Markdown document of approximately
// the requested byte size. The output mixes headers, prose paragraphs, fenced
// code blocks, tables, bullet lists, links, and scattered emoji -- including
// ZWJ sequences, variation selectors, and skin-tone modifiers -- to simulate
// a complex, real-world document that the library must process efficiently.
func buildLargeMarkdown(targetBytes int) string {
	// Reusable fragments that appear in real Markdown documents.
	const header = "# Project Status Report\n\n"
	const subheader = "## Section %d: Feature Update\n\n"
	const prose = "The deployment pipeline completed successfully and all integration " +
		"tests passed on the staging environment. Performance metrics remain within " +
		"acceptable thresholds across all monitored endpoints. The team reviewed the " +
		"architecture decision records and agreed on the proposed changes to the " +
		"caching layer.\n\n"
	const proseEmoji = "Status: \u2705 passing | Build: \U0001F680 deployed | " +
		"Coverage: \U0001F4CA 94%% | Alerts: \u26A0\uFE0F none | " +
		"Review: \U0001F44D\U0001F3FD approved | " + // skin-tone modifier
		"Team: \U0001F468\u200D\U0001F4BB\u200D\U0001F469\u200D\U0001F4BB pair | " + // ZWJ
		"Flag: \U0001F1FA\U0001F1F8 | " + // flag sequence
		"Heart: \u2764\uFE0F\u200D\U0001F525\n\n" // variation selector + ZWJ
	const codeBlock = "```go\nfunc process(ctx context.Context, items []Item) error {\n" +
		"\tfor _, item := range items {\n" +
		"\t\tif err := validate(item); err != nil {\n" +
		"\t\t\treturn fmt.Errorf(\"validate %s: %w\", item.ID, err)\n" +
		"\t\t}\n" +
		"\t}\n" +
		"\treturn nil\n}\n```\n\n"
	const table = "| Metric | Value | Status |\n" +
		"|--------|-------|--------|\n" +
		"| Latency p99 | 142ms | \u2705 |\n" +
		"| Error rate | 0.02%% | \u2705 |\n" +
		"| Throughput | 12,400 req/s | \U0001F680 |\n" +
		"| Memory | 1.2 GiB | \u26A0\uFE0F |\n\n"
	const bulletList = "- Refactored the authentication middleware \U0001F512\n" +
		"- Added retry logic with exponential backoff \u23F3\n" +
		"- Updated API documentation for v2 endpoints \U0001F4DD\n" +
		"- Resolved flaky test in CI \U0001F527\n" +
		"- Deployed canary to 5%% of traffic \U0001F6A6\n\n"
	const link = "[Architecture Decision Records](https://example.com/adr) | " +
		"[Runbook](https://example.com/runbook) | " +
		"[Dashboard](https://example.com/dashboard)\n\n"
	const whitespaceMessy = "Some   text   with   extra   spaces   scattered   " +
		"throughout   the   document.\n\n\n\n" +
		"And   multiple   blank   lines   above.\n\n"

	var b strings.Builder
	b.Grow(targetBytes + 4096) // pre-allocate to reduce resizing
	b.WriteString(header)
	section := 0
	for b.Len() < targetBytes {
		section++
		b.WriteString(fmt.Sprintf(subheader, section))
		b.WriteString(prose)
		b.WriteString(fmt.Sprintf(proseEmoji /* %% are literal */))
		b.WriteString(codeBlock)
		b.WriteString(fmt.Sprintf(table /* %% are literal */))
		b.WriteString(bulletList)
		b.WriteString(link)
		b.WriteString(whitespaceMessy)
	}
	return b.String()
}

// largeSizes defines the document sizes used by scaled benchmarks.
// Each entry is a human-readable label paired with a target byte count.
var largeSizes = []struct {
	name  string
	bytes int
}{
	{"10KB", 10 * 1024},
	{"100KB", 100 * 1024},
	{"1MB", 1024 * 1024},
}

// ---------- Original benchmarks (unchanged) ----------

func BenchmarkDemojify(b *testing.B) {
	input := benchInput(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.Demojify(input)
	}
}

func BenchmarkContainsEmoji(b *testing.B) {
	input := benchInput(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.ContainsEmoji(input)
	}
}

func BenchmarkNormalize(b *testing.B) {
	input := strings.Repeat("Hello   World  \n\n\n\nMore text  here\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.Normalize(input)
	}
}

func BenchmarkReplace(b *testing.B) {
	input := benchInput(100)
	repl := demojify.DefaultReplacements()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.Replace(input, repl)
	}
}

func BenchmarkReplaceCount(b *testing.B) {
	input := benchInput(100)
	repl := demojify.DefaultReplacements()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.ReplaceCount(input, repl)
	}
}

func BenchmarkSanitize(b *testing.B) {
	input := benchInput(100)
	opts := demojify.DefaultOptions()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.Sanitize(input, opts)
	}
}

func BenchmarkScanDir(b *testing.B) {
	dir := b.TempDir()
	// Create 10 files with emoji content.
	for i := 0; i < 10; i++ {
		name := "file" + strings.Repeat("x", i) + ".md"
		content := benchInput(10)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			b.Fatalf("write %s: %v", name, err)
		}
	}
	cfg := demojify.DefaultScanConfig()
	cfg.Root = dir
	cfg.ExemptSuffixes = nil // don't skip anything
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := demojify.ScanDir(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindAll(b *testing.B) {
	input := benchInput(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demojify.FindAll(input)
	}
}

// ---------- Large-document benchmarks ----------
//
// Each benchmark generates a realistic Markdown document at multiple sizes
// (10 KB, 100 KB, 1 MB) and reports bytes-per-op so throughput can be
// compared across runs.

func BenchmarkLargeDemojify(b *testing.B) {
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.Demojify(input)
			}
		})
	}
}

func BenchmarkLargeContainsEmoji(b *testing.B) {
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.ContainsEmoji(input)
			}
		})
	}
}

func BenchmarkLargeNormalize(b *testing.B) {
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.Normalize(input)
			}
		})
	}
}

func BenchmarkLargeReplace(b *testing.B) {
	repl := demojify.DefaultReplacements()
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.Replace(input, repl)
			}
		})
	}
}

func BenchmarkLargeReplaceCount(b *testing.B) {
	repl := demojify.DefaultReplacements()
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.ReplaceCount(input, repl)
			}
		})
	}
}

func BenchmarkLargeSanitize(b *testing.B) {
	opts := demojify.DefaultOptions()
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.Sanitize(input, opts)
			}
		})
	}
}

func BenchmarkLargeFindAll(b *testing.B) {
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.FindAll(input)
			}
		})
	}
}

func BenchmarkLargeFindAllMapped(b *testing.B) {
	repl := demojify.DefaultReplacements()
	for _, sz := range largeSizes {
		input := buildLargeMarkdown(sz.bytes)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				demojify.FindAllMapped(input, repl)
			}
		})
	}
}

// BenchmarkLargeScanDir creates a directory tree of large Markdown files and
// benchmarks the full scan pipeline at different total-data sizes.
func BenchmarkLargeScanDir(b *testing.B) {
	for _, sz := range largeSizes {
		b.Run(sz.name, func(b *testing.B) {
			dir := b.TempDir()
			// Distribute the target size across 5 files to simulate a
			// realistic directory with multiple documents.
			const fileCount = 5
			perFile := sz.bytes / fileCount
			totalWritten := 0
			for i := 0; i < fileCount; i++ {
				content := buildLargeMarkdown(perFile)
				totalWritten += len(content)
				name := fmt.Sprintf("doc_%02d.md", i)
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					b.Fatalf("write %s: %v", name, err)
				}
			}
			cfg := demojify.DefaultScanConfig()
			cfg.Root = dir
			cfg.ExemptSuffixes = nil
			b.SetBytes(int64(totalWritten))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := demojify.ScanDir(cfg); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLargeScanFile benchmarks scanning a single large file.
func BenchmarkLargeScanFile(b *testing.B) {
	opts := demojify.DefaultOptions()
	for _, sz := range largeSizes {
		b.Run(sz.name, func(b *testing.B) {
			dir := b.TempDir()
			content := buildLargeMarkdown(sz.bytes)
			path := filepath.Join(dir, "large.md")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				b.Fatalf("write: %v", err)
			}
			b.SetBytes(int64(len(content)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := demojify.ScanFile(path, opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLargeFindMatchesInFile benchmarks per-occurrence match extraction
// from a single large file.
func BenchmarkLargeFindMatchesInFile(b *testing.B) {
	repl := demojify.DefaultReplacements()
	for _, sz := range largeSizes {
		b.Run(sz.name, func(b *testing.B) {
			dir := b.TempDir()
			content := buildLargeMarkdown(sz.bytes)
			path := filepath.Join(dir, "large.md")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				b.Fatalf("write: %v", err)
			}
			b.SetBytes(int64(len(content)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := demojify.FindMatchesInFile(path, repl); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLargeSanitizeFile benchmarks the full sanitize-and-write-back path
// on a single large file. Each iteration re-creates the file so the write
// path is exercised every time.
func BenchmarkLargeSanitizeFile(b *testing.B) {
	opts := demojify.DefaultOptions()
	for _, sz := range largeSizes {
		b.Run(sz.name, func(b *testing.B) {
			dir := b.TempDir()
			content := buildLargeMarkdown(sz.bytes)
			path := filepath.Join(dir, "large.md")
			b.SetBytes(int64(len(content)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Re-write the dirty file before each sanitize pass.
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					b.Fatalf("write: %v", err)
				}
				if _, err := demojify.SanitizeFile(path, opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLargeReplaceFile benchmarks file-level emoji substitution on a
// large file. Each iteration re-creates the dirty file.
func BenchmarkLargeReplaceFile(b *testing.B) {
	repl := demojify.DefaultReplacements()
	for _, sz := range largeSizes {
		b.Run(sz.name, func(b *testing.B) {
			dir := b.TempDir()
			content := buildLargeMarkdown(sz.bytes)
			path := filepath.Join(dir, "large.md")
			b.SetBytes(int64(len(content)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					b.Fatalf("write: %v", err)
				}
				if _, err := demojify.ReplaceFile(path, repl); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
