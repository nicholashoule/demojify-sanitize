package demojify_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func ExampleDemojify() {
	fmt.Println(demojify.Demojify("\U0001F680 Deploy complete! Check the logs \U0001F4CA"))
	// Output:
	//  Deploy complete! Check the logs
}

func ExampleContainsEmoji() {
	fmt.Println(demojify.ContainsEmoji("Hello \U0001F600 World"))
	fmt.Println(demojify.ContainsEmoji("Hello World"))
	// Output:
	// true
	// false
}

func ExampleNormalize() {
	fmt.Println(demojify.Normalize("Hello   World\n\n\nMore text"))
	// Output:
	// Hello World
	//
	// More text
}

func ExampleSanitize() {
	input := "\U0001F680 Deploy complete!\n\n\nCheck the logs \U0001F4CA"
	fmt.Println(demojify.Sanitize(input, demojify.DefaultOptions()))
	// Output:
	// Deploy complete!
	//
	// Check the logs
}

func ExampleSanitize_selective() {
	// Only remove emojis, leave whitespace untouched.
	opts := demojify.Options{RemoveEmojis: true}
	fmt.Println(demojify.Sanitize("Sure! \U0001F389 Done.", opts))
	// Output:
	// Sure!  Done.
}

// ExampleContainsEmoji_contentGate shows how to use ContainsEmoji as a
// guard before persisting or forwarding user-submitted text.
func ExampleContainsEmoji_contentGate() {
	report := "Q3 results: Revenue up 12% \U0001F4C8"

	if demojify.ContainsEmoji(report) {
		// Strip emojis and normalize before storing.
		opts := demojify.Options{RemoveEmojis: true, NormalizeWhitespace: true}
		fmt.Println(demojify.Sanitize(report, opts))
	}
	// Output:
	// Q3 results: Revenue up 12%
}

// ExampleSanitize_httpHandler shows how to wrap Sanitize in an HTTP handler
// that cleans an incoming plain-text request body before processing it.
// This example is compiled but not executed (no Output comment).
func ExampleSanitize_httpHandler() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		clean := demojify.Sanitize(string(body), demojify.DefaultOptions())
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, clean)
	})
	_ = handler
}

// ExampleSanitize_markdownFiles shows how to sanitize a set of Markdown
// files in place -- suitable for a pre-commit hook or CI step.
// This example is compiled but not executed (no Output comment).
func ExampleSanitize_markdownFiles() {
	paths := []string{"README.md", "CHANGELOG.md", "CONTRIBUTING.md"}
	opts := demojify.DefaultOptions()

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			log.Printf("read %s: %v", p, err)
			continue
		}
		clean := demojify.Sanitize(string(data), opts)
		if err := os.WriteFile(p, []byte(clean), 0o644); err != nil {
			log.Printf("write %s: %v", p, err)
		}
	}
}
