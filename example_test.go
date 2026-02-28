package demojify_test

import (
	"fmt"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

func ExampleDemojify() {
	fmt.Println(demojify.Demojify("🚀 Deploy complete! Check the logs 📊"))
	// Output:
	//  Deploy complete! Check the logs
}

func ExampleContainsEmoji() {
	fmt.Println(demojify.ContainsEmoji("Hello 😀 World"))
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
	input := "Certainly!\n🚀 Deploy complete!\n\n\nCheck the logs 📊"
	fmt.Println(demojify.Sanitize(input, demojify.DefaultOptions()))
	// Output:
	// Deploy complete!
	//
	// Check the logs
}

func ExampleSanitize_selective() {
	// Only remove emojis, leave whitespace and AI clutter untouched.
	opts := demojify.Options{RemoveEmojis: true}
	fmt.Println(demojify.Sanitize("Sure! 🎉 Done.", opts))
	// Output:
	// Sure!  Done.
}
