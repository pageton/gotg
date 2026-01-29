package parsemode

import (
	"strconv"
	"strings"
)

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

		if content == "" {
			break
		}

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
	delim := "__"
	delimLen := len(delim)

	for {
		start := strings.Index(markdown, delim)
		if start == -1 {
			break
		}

		if start > 0 && markdown[start-1] == '*' {
			markdown = markdown[:start] + delim + markdown[start+delimLen:]
			continue
		}

		rest := markdown[start+delimLen:]
		end := strings.Index(rest, delim)
		if end == -1 {
			break
		}

		content := rest[:end]

		if end+delimLen < len(rest) && rest[end+delimLen] == '*' {
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
	for {
		start := strings.Index(markdown, "[")
		if start == -1 {
			break
		}

		textEnd := strings.Index(markdown[start:], "]")
		if textEnd == -1 {
			break
		}
		textEnd += start

		if textEnd+1 >= len(markdown) || markdown[textEnd+1] != '(' {
			markdown = markdown[:textEnd] + markdown[textEnd+1:]
			continue
		}

		urlEnd := strings.Index(markdown[textEnd+2:], ")")
		if urlEnd == -1 {
			break
		}
		urlEnd += textEnd + 2

		text := markdown[start+1 : textEnd]
		url := markdown[textEnd+2 : urlEnd]

		if url == "" {
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

		if _, err := strconv.ParseInt(emojiID, 10, 64); err != nil {
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
	escapeChars := []string{"*", "_", "~", "|", "`", "[", "]", "(", ")", "{", "}", "<", ">", "!", ".", "-", "=", "+"}

	result := input
	for _, ch := range escapeChars {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}

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
