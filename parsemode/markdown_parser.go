package parsemode

import (
	"strconv"
	"strings"
)

// MarkdownParser implements the Parser interface for MarkdownV2 formatting.
// It is thread-safe and can be used concurrently.
type MarkdownParser struct {
	htmlParser *HTMLParser
}

// NewMarkdownParser creates a new Markdown parser instance.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		htmlParser: NewHTMLParser(),
	}
}

// Parse parses MarkdownV2 formatted text and returns the result with Telegram entities.
func (p *MarkdownParser) Parse(input string) (*ParseResult, error) {
	if input == "" {
		return &ParseResult{Text: "", Entities: nil}, nil
	}

	html := p.MarkdownToHTML(input)
	return p.htmlParser.Parse(html)
}

// MarkdownToHTML converts MarkdownV2 syntax to HTML.
func (p *MarkdownParser) MarkdownToHTML(markdown string) string {
	if markdown == "" {
		return ""
	}

	markdown, placeholders := p.handleEscapes(markdown)
	markdown = p.convertCodeBlockSyntax(markdown)
	markdown = p.convertCodeSyntax(markdown)

	for _, conv := range [][3]string{
		{"**", "b", "b"},
		{"__", "b", "b"},
		{"*", "i", "i"},
		{"_", "i", "i"},
		{"~", "s", "s"},
		{"~~", "s", "s"},
		{"||", "spoiler", "spoiler"},
	} {
		markdown = p.convertSyntax(markdown, conv[0], conv[1], conv[2])
	}

	markdown = p.convertUnderlineSyntax(markdown)
	markdown = p.convertLinksSyntax(markdown)
	markdown = p.convertEmojiSyntax(markdown)
	markdown = p.convertBlockquoteSyntax(markdown)
	markdown = p.restoreEscapes(markdown, placeholders)

	return strings.TrimSpace(markdown)
}

// handleEscapes preserves escaped characters by replacing them with placeholders.
func (p *MarkdownParser) handleEscapes(markdown string) (string, map[string]string) {
	escapeChars := []string{"*", "_", "~", "|", "`", "[", "]", "(", ")", "{", "}", "<", ">", "!"}
	placeholders := make(map[string]string, len(escapeChars))

	for i, ch := range escapeChars {
		esc := "\\" + ch
		placeholder := "\x00ESC" + strconv.Itoa(i) + "\x00"
		placeholders[placeholder] = ch
		markdown = strings.ReplaceAll(markdown, esc, placeholder)
	}

	return markdown, placeholders
}

// restoreEscapes restores the escaped characters from placeholders.
func (p *MarkdownParser) restoreEscapes(markdown string, placeholders map[string]string) string {
	for placeholder, ch := range placeholders {
		markdown = strings.ReplaceAll(markdown, placeholder, ch)
	}
	return markdown
}
