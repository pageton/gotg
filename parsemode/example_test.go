package parsemode

import (
	"fmt"

	"github.com/gotd/td/tg"
)

// ExampleHTMLParser demonstrates basic HTML parsing.
func ExampleHTMLParser() {
	parser := NewHTMLParser()

	// Simple bold text
	result, _ := parser.Parse("<b>Hello, World!</b>")
	fmt.Println(result.Text)
	// Mixed formatting
	result, _ = parser.Parse("<b>Bold</b> and <i>italic</i> text")
	fmt.Println(result.Text)
	// Link
	result, _ = parser.Parse(`<a href="https://example.com">Click here</a>`)
	fmt.Println(result.Text)
	// Output:
	// Hello, World!
	// Bold and italic text
	// Click here
}

// ExampleHTMLParser_code demonstrates code parsing.
func ExampleHTMLParser_code() {
	parser := NewHTMLParser()

	// Inline code
	result, _ := parser.Parse("<code>fmt.Println()</code>")
	fmt.Println(result.Text)

	// Code block with language
	result, _ = parser.Parse(`<pre class="language-go">func main() {}</pre>`)
	fmt.Println(result.Text)

	pre := result.Entities[0].(*tg.MessageEntityPre)
	fmt.Println(pre.Language)
	// Output:
	// fmt.Println()
	// func main() {}
	// go
}

// ExampleHTMLParser_spoiler demonstrates spoiler parsing.
func ExampleHTMLParser_spoiler() {
	parser := NewHTMLParser()

	result, _ := parser.Parse("<spoiler>Secret message</spoiler>")
	fmt.Println(result.Text)
	// Output: Secret message
}

// ExampleHTMLParser_blockquote demonstrates blockquote parsing.
func ExampleHTMLParser_blockquote() {
	parser := NewHTMLParser()

	// Regular blockquote
	result, _ := parser.Parse("<blockquote>This is a quote</blockquote>")
	fmt.Println(result.Text)

	// Collapsed blockquote
	result, _ = parser.Parse(`<blockquote collapsed="true">Hidden content</blockquote>`)
	fmt.Println(result.Text)
	// Output:
	// This is a quote
	// Hidden content
}

// ExampleHTMLParser_customEmoji demonstrates custom emoji parsing.
func ExampleHTMLParser_customEmoji() {
	parser := NewHTMLParser()

	result, _ := parser.Parse(`<emoji id="12345">🎉</emoji>`)
	fmt.Println(result.Text)
	// Output: 🎉
}

// ExampleMarkdownParser_bold demonstrates bold text in Markdown.
func ExampleMarkdownParser_bold() {
	parser := NewMarkdownParser()

	// Double asterisk
	result, _ := parser.Parse("**Bold text**")
	fmt.Println(result.Text)

	// Double underscore
	result, _ = parser.Parse("__Also bold__")
	fmt.Println(result.Text)
	// Output:
	// Bold text
	// Also bold
}

// ExampleMarkdownParser_italic demonstrates italic text in Markdown.
func ExampleMarkdownParser_italic() {
	parser := NewMarkdownParser()

	// Single asterisk
	result, _ := parser.Parse("*Italic text*")
	fmt.Println(result.Text)

	// Single underscore
	result, _ = parser.Parse("_Also italic_")
	fmt.Println(result.Text)
	// Output:
	// Italic text
	// Also italic
}

// ExampleMarkdownParser_code demonstrates code parsing in Markdown.
func ExampleMarkdownParser_code() {
	parser := NewMarkdownParser()

	// Inline code
	result, _ := parser.Parse("`code`")
	fmt.Println(result.Text)

	// Code block
	result, _ = parser.Parse("```\ncode block\n```")
	fmt.Println(result.Text)

	// Code block with language
	result, _ = parser.Parse("```go\nfunc main() {}\n```")
	fmt.Println(result.Text)
	// Output:
	// code
	// code block
	// func main() {}
}

// ExampleMarkdownParser_link demonstrates link parsing in Markdown.
func ExampleMarkdownParser_link() {
	parser := NewMarkdownParser()

	result, _ := parser.Parse("[Open Google](https://google.com)")
	fmt.Println(result.Text)
	// Output: Open Google
}

// ExampleMarkdownParser_spoiler demonstrates spoiler parsing in Markdown.
func ExampleMarkdownParser_spoiler() {
	parser := NewMarkdownParser()

	result, _ := parser.Parse("||Spoiler content||")
	fmt.Println(result.Text)
	// Output: Spoiler content
}

// ExampleMarkdownParser_blockquote demonstrates blockquote parsing in Markdown.
func ExampleMarkdownParser_blockquote() {
	parser := NewMarkdownParser()

	// Regular blockquote
	result, _ := parser.Parse("> This is a quote")
	fmt.Println(result.Text)

	// Collapsed blockquote
	result, _ = parser.Parse(">>> Hidden content")
	fmt.Println(result.Text)
	// Output:
	// This is a quote
	// Hidden content
}

// ExampleMarkdownParser_customEmoji demonstrates custom emoji parsing in Markdown.
func ExampleMarkdownParser_customEmoji() {
	parser := NewMarkdownParser()

	result, _ := parser.Parse("::12345::")
	fmt.Printf("Text: '%s', Emoji ID present: %v\n", result.Text, len(result.Entities) > 0)
	// Output: Text: '', Emoji ID present: true
}

// ExampleParseMode demonstrates ParseMode usage.
func ExampleParseMode() {
	// Different parse modes
	modes := []ParseMode{
		ModeNone,
		ModeHTML,
		ModeMarkdown,
	}

	for _, mode := range modes {
		fmt.Printf("Mode: %s, Valid: %v\n", mode, mode.IsValid())
	}
	// Output:
	// Mode: , Valid: true
	// Mode: HTML, Valid: true
	// Mode: MarkdownV2, Valid: true
}

// ExampleEntityBuilder demonstrates EntityBuilder usage.
func ExampleEntityBuilder() {
	builder := NewEntityBuilder()

	// Add entities
	builder.Add(&tg.MessageEntityBold{Offset: 0, Length: 4})
	builder.Add(&tg.MessageEntityItalic{Offset: 5, Length: 6})

	entities := builder.Build()
	fmt.Printf("Built %d entities\n", len(entities))

	// Clone the builder
	cloned := builder.Clone()
	fmt.Printf("Cloned builder has %d entities\n", len(cloned.Build()))

	// Reset
	builder.Reset()
	fmt.Printf("After reset: %d entities\n", len(builder.Build()))
	// Output:
	// Built 2 entities
	// Cloned builder has 2 entities
	// After reset: 0 entities
}

// ExampleUTF16RuneCountInString demonstrates UTF-16 counting.
func ExampleUTF16RuneCountInString() {
	// ASCII strings have the same length in UTF-8 and UTF-16
	ascii := "hello"
	fmt.Printf("ASCII: %d\n", UTF16RuneCountInString(ascii))

	// Emoji require 2 UTF-16 code units (surrogate pair)
	emoji := "🎉"
	fmt.Printf("Emoji: %d\n", UTF16RuneCountInString(emoji))
	// Output:
	// ASCII: 5
	// Emoji: 2
}

// ExampleFormatEntity demonstrates FormatEntity usage.
func ExampleFormatEntity() {
	// Format bold text
	html := FormatEntity(string(EntityTypeBold), "hello", nil)
	fmt.Println(html)

	// Format code with language
	html = FormatEntity(string(EntityTypePre), "code", map[string]string{"language": "go"})
	fmt.Println(html)

	// Format link
	html = FormatEntity(string(EntityTypeTextURL), "Click here", map[string]string{"url": "https://example.com"})
	fmt.Println(html)
	// Output:
	// <b>hello</b>
	// <pre class="language-go">code</pre>
	// <a href="https://example.com">Click here</a>
}
