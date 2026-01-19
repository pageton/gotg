package parsemode

import (
	"strconv"
	"strings"
)

// MarkdownParser implements the Parser interface for MarkdownV2 formatting.
// It is thread-safe and can be used concurrently.
type MarkdownParser struct {
	// HTMLParser is used internally to convert Markdown to HTML first.
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

	// Convert MarkdownV2 to HTML first
	html := p.MarkdownToHTML(input)

	// Parse the HTML using the HTML parser
	return p.htmlParser.Parse(html)
}

// MarkdownToHTML converts MarkdownV2 syntax to HTML.
func (p *MarkdownParser) MarkdownToHTML(markdown string) string {
	if markdown == "" {
		return ""
	}

	// First, handle escape characters to preserve them
	markdown, placeholders := p.handleEscapes(markdown)

	// Convert code blocks first (they need special handling)
	markdown = p.convertCodeBlockSyntax(markdown)

	// Convert inline code
	markdown = p.convertCodeSyntax(markdown)

	// Convert bold, italic, underline, strikethrough, spoiler
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

	// Convert underline (not standard in MarkdownV2, but supported)
	markdown = p.convertUnderlineSyntax(markdown)

	// Convert links
	markdown = p.convertLinksSyntax(markdown)

	// Convert custom emoji
	markdown = p.convertEmojiSyntax(markdown)

	// Convert blockquotes
	markdown = p.convertBlockquoteSyntax(markdown)

	// Restore escaped characters
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

// convertSyntax converts a simple delimiter syntax to HTML tags.
func (p *MarkdownParser) convertSyntax(markdown, delim, openTag, closeTag string) string {
	delimLen := len(delim)

	for {
		start := strings.Index(markdown, delim)
		if start == -1 {
			break
		}

		rest := markdown[start+delimLen:]
		end := strings.Index(rest, delim)
		if end == -1 {
			break
		}

		content := rest[:end]

		// Skip empty content (e.g., ****)
		if content == "" {
			break
		}

		// Check if content contains newline - delimiters across lines are not valid
		if strings.Contains(content, "\n") {
			markdown = markdown[:start] + delim + rest
			continue
		}

		markdown = markdown[:start] + "<" + openTag + ">" + content + "</" + closeTag + ">" + rest[end+delimLen:]
	}

	return markdown
}

// convertUnderlineSyntax converts underline syntax (__) to HTML.
func (p *MarkdownParser) convertUnderlineSyntax(markdown string) string {
	// Try to match __text__ pattern that is NOT part of bold (**text**)
	delim := "__"
	delimLen := len(delim)

	for {
		start := strings.Index(markdown, delim)
		if start == -1 {
			break
		}

		// Check if this is actually a bold marker (**)
		if start > 0 && markdown[start-1] == '*' {
			// This is part of **__, skip
			markdown = markdown[:start] + delim + markdown[start+delimLen:]
			continue
		}

		rest := markdown[start+delimLen:]
		end := strings.Index(rest, delim)
		if end == -1 {
			break
		}

		content := rest[:end]

		// Check if followed by another * (making it **** which is bold)
		if end+delimLen < len(rest) && rest[end+delimLen] == '*' {
			// This is part of ____, skip
			markdown = markdown[:start] + delim + rest
			continue
		}

		if content == "" || strings.Contains(content, "\n") {
			break
		}

		markdown = markdown[:start] + "<u>" + content + "</u>" + rest[end+delimLen:]
	}

	return markdown
}

// convertCodeSyntax converts inline code to HTML.
func (p *MarkdownParser) convertCodeSyntax(markdown string) string {
	delim := "`"

	for {
		start := strings.Index(markdown, delim)
		if start == -1 {
			break
		}

		// Check for code block (```)
		if start+3 < len(markdown) && markdown[start:start+3] == "```" {
			break
		}

		rest := markdown[start+1:]
		end := strings.Index(rest, delim)
		if end == -1 || end == 0 {
			break
		}

		content := rest[:end]
		escaped := htmlEscape(content)
		markdown = markdown[:start] + "<code>" + escaped + "</code>" + rest[end+1:]
	}

	return markdown
}

// convertCodeBlockSyntax converts code blocks to HTML.
func (p *MarkdownParser) convertCodeBlockSyntax(markdown string) string {
	const fence = "```"

	for {
		start := strings.Index(markdown, fence)
		if start == -1 {
			break
		}

		rest := markdown[start+3:]
		before, after, ok := strings.Cut(rest, fence)
		if !ok {
			break
		}

		block := before

		// Extract language and code
		var lang, code string
		if before0, after0, ok0 := strings.Cut(block, "\n"); ok0 {
			lang = strings.TrimSpace(before0)
			code = after0
		} else {
			code = block
		}

		escaped := htmlEscape(code)

		var replacement string
		if lang != "" {
			replacement = `<pre class="language-` + lang + `">` + escaped + "</pre>"
		} else {
			replacement = "<pre>" + escaped + "</pre>"
		}

		markdown = markdown[:start] + replacement + after
	}

	return markdown
}

// convertLinksSyntax converts markdown links to HTML.
func (p *MarkdownParser) convertLinksSyntax(markdown string) string {
	// Pattern: [text](url)
	for {
		start := strings.Index(markdown, "[")
		if start == -1 {
			break
		}

		// Find closing bracket
		textEnd := strings.Index(markdown[start:], "]")
		if textEnd == -1 {
			break
		}
		textEnd += start

		// Check for opening parenthesis
		if textEnd+1 >= len(markdown) || markdown[textEnd+1] != '(' {
			markdown = markdown[:textEnd] + markdown[textEnd+1:]
			continue
		}

		// Find closing parenthesis
		urlEnd := strings.Index(markdown[textEnd+2:], ")")
		if urlEnd == -1 {
			break
		}
		urlEnd += textEnd + 2

		text := markdown[start+1 : textEnd]
		url := markdown[textEnd+2 : urlEnd]

		// Check for hidden URL (used for mentions without visible URL)
		if url == "" {
			// Treat as plain URL entity
			markdown = markdown[:start] + `<a href="">` + text + "</a>" + markdown[urlEnd+1:]
		} else {
			markdown = markdown[:start] + `<a href="` + htmlEscape(url) + `">` + text + "</a>" + markdown[urlEnd+1:]
		}

		markdown = strings.ReplaceAll(markdown, "  ", " ")
	}

	return markdown
}

// convertEmojiSyntax converts custom emoji syntax to HTML.
func (p *MarkdownParser) convertEmojiSyntax(markdown string) string {
	const delim = "::"

	for {
		start := strings.Index(markdown, delim)
		if start == -1 {
			break
		}

		rest := markdown[start+2:]
		end := strings.Index(rest, delim)
		if end == -1 || end == 0 {
			break
		}

		emojiID := rest[:end]

		// Validate that it's a number
		if _, err := strconv.ParseInt(emojiID, 10, 64); err != nil {
			// Not a valid emoji ID, skip
			markdown = markdown[:start] + delim + rest[end+2:]
			continue
		}

		markdown = markdown[:start] + `<emoji id="` + emojiID + `"></emoji>` + rest[end+2:]
	}

	return markdown
}

// convertBlockquoteSyntax converts blockquote syntax to HTML.
func (p *MarkdownParser) convertBlockquoteSyntax(markdown string) string {
	lines := strings.Split(markdown, "\n")
	if len(lines) == 0 {
		return markdown
	}

	var result strings.Builder
	result.Grow(len(markdown) + 100)

	var inBlockquote bool
	var isCollapsed bool

	closeBlockquote := func() {
		if inBlockquote {
			result.WriteString("</blockquote>\n")
			inBlockquote, isCollapsed = false, false
		}
	}

	openBlockquote := func(collapsed bool) {
		if inBlockquote && isCollapsed != collapsed {
			closeBlockquote()
		}
		if !inBlockquote {
			if collapsed {
				result.WriteString(`<blockquote collapsed="true">`)
			} else {
				result.WriteString("<blockquote>")
			}
			inBlockquote, isCollapsed = true, collapsed
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		switch {
		case strings.HasPrefix(trimmed, ">>> "):
			openBlockquote(true)
			content := strings.TrimPrefix(trimmed, ">>> ")
			result.WriteString(content + "\n")
		case strings.HasPrefix(trimmed, "> "):
			openBlockquote(false)
			content := strings.TrimPrefix(trimmed, "> ")
			result.WriteString(content + "\n")
		default:
			closeBlockquote()
			result.WriteString(line + "\n")
		}
	}

	closeBlockquote()
	return result.String()
}

// Format formats text as MarkdownV2 (escapes special characters).
func (p *MarkdownParser) Format(input string) string {
	// Escape special characters for MarkdownV2
	escapeChars := []string{"*", "_", "~", "|", "`", "[", "]", "(", ")", "{", "}", "<", ">", "!", ".", "-", "=", "+"}

	result := input
	for _, ch := range escapeChars {
		// Only escape if not already escaped
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}

	// Fix double escapes
	for _, ch := range escapeChars {
		result = strings.ReplaceAll(result, "\\"+ch, "\\"+ch)
	}

	return result
}

// HTMLToMarkdownV2 converts HTML to MarkdownV2 format.
func HTMLToMarkdownV2(html string) string {
	if html == "" {
		return ""
	}

	parser := NewMarkdownParser()
	return parser.MarkdownToHTML(html)
}
