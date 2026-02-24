package parsemode

import (
	"net/url"
	"strings"
)

// isValidTelegramURL validates that a URL protocol is safe for Telegram use.
// Returns true for safe protocols (http, https, tg, mailto) and empty hrefs.
// Returns false for dangerous protocols (javascript, data, vbscript, file).
func isValidTelegramURL(href string) bool {
	if href == "" {
		return true
	}
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	switch u.Scheme {
	case "http", "https", "tg", "mailto":
		return true
	case "javascript", "data", "vbscript", "file":
		return false
	case "":
		// No scheme - could be a protocol-relative URL or domain-only URL
		// (e.g., "t.me/username" or "//example.com/path")
		return true
	default:
		return false
	}
}

// normalizeURL ensures a URL has a proper scheme for Telegram.
// Protocol-less URLs like "t.me/username" get "https://" prepended.
func normalizeURL(href string) string {
	if href == "" {
		return href
	}
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	if u.Scheme == "" && u.Host == "" && !strings.HasPrefix(href, "//") {
		// Looks like "t.me/username" - no scheme and parsed as path only
		// Prepend https:// to make it a valid URL
		return "https://" + href
	}
	if strings.HasPrefix(href, "//") {
		// Protocol-relative URL like "//example.com/path"
		return "https:" + href
	}
	return href
}

// HTMLParser implements the Parser interface for HTML formatting.
// It is thread-safe and can be used concurrently.
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser instance.
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// Parse parses HTML formatted text and returns the result with Telegram entities.
func (p *HTMLParser) Parse(input string) (*ParseResult, error) {
	if input == "" {
		return &ParseResult{Text: "", Entities: nil}, nil
	}

	cleanedText, tags, err := p.parseHTMLToTags(input)
	if err != nil {
		return &ParseResult{Text: input, Entities: nil}, nil
	}

	entities := p.parseTagsToEntities(tags)

	return &ParseResult{
		Text:     cleanedText,
		Entities: entities,
	}, nil
}

// htmlToken represents a token from HTML parsing.
type htmlToken struct {
	isTag     bool
	isClosing bool
	tagName   string
	attrs     map[string]string
	text      string
}

// simpleHTMLTokenize tokenizes HTML input into a stream of tokens.
func simpleHTMLTokenize(html string) []htmlToken {
	var tokens []htmlToken
	i := 0

	for i < len(html) {
		if html[i] == '<' {
			tagEnd := i + 1
			for tagEnd < len(html) && html[tagEnd] != '>' {
				tagEnd++
			}

			if tagEnd >= len(html) {
				tokens = append(tokens, htmlToken{isTag: false, text: html[i:]})
				break
			}

			tagContent := html[i+1 : tagEnd]
			isClosing := strings.HasPrefix(tagContent, "/")
			if isClosing {
				tagContent = tagContent[1:]
			}

			parts := strings.Fields(tagContent)
			if len(parts) > 0 {
				tagName := strings.ToLower(parts[0])
				attrs := make(map[string]string)

				for _, part := range parts[1:] {
					if strings.Contains(part, "=") {
						kv := strings.SplitN(part, "=", 2)
						key := strings.ToLower(kv[0])
						value := strings.Trim(kv[1], "\"'")
						attrs[key] = value
					} else {
						attrs[part] = "true"
					}
				}

				tokens = append(tokens, htmlToken{
					isTag:     true,
					isClosing: isClosing,
					tagName:   tagName,
					attrs:     attrs,
					text:      html[i : tagEnd+1],
				})
			} else {
				tokens = append(tokens, htmlToken{
					isTag: false,
					text:  html[i : tagEnd+1],
				})
			}
			i = tagEnd + 1
		} else {
			textStart := i
			for i < len(html) && html[i] != '<' {
				i++
			}
			tokens = append(tokens, htmlToken{
				isTag: false,
				text:  html[textStart:i],
			})
		}
	}

	return tokens
}

// tag represents a parsed HTML tag with position information.
type tag struct {
	Type      string
	Offset    int32
	Length    int32
	Attrs     map[string]string
	HasNested bool
}

// parseHTMLToTags converts HTML string into cleaned text and tag information.
func (p *HTMLParser) parseHTMLToTags(htmlStr string) (string, []tag, error) {
	tokens := simpleHTMLTokenize(htmlStr)

	var textBuf strings.Builder
	var tagOffsets []tag

	type openTag struct {
		tag    tag
		tagIdx int
	}
	var openTags []openTag

	for _, token := range tokens {
		if !token.isTag { //nolint:gocritic // ifElseChain: conditions involve struct fields, not suitable for switch
			textBuf.WriteString(htmlUnescape(token.text))
		} else if !token.isClosing && supportedTag(token.tagName) {
			currentOffset := utf16RuneCountInString(textBuf.String())
			newTag := tag{
				Type:   token.tagName,
				Offset: currentOffset,
				Attrs:  token.attrs,
			}
			tagIdx := len(tagOffsets)
			tagOffsets = append(tagOffsets, newTag)
			openTags = append(openTags, openTag{tag: newTag, tagIdx: tagIdx})
		} else if token.isClosing {
			matched := false
			tagName := strings.ToLower(token.tagName)
			for i := len(openTags) - 1; i >= 0; i-- {
				if openTags[i].tag.Type == tagName {
					currentOffset := utf16RuneCountInString(textBuf.String())
					tagOffsets[openTags[i].tagIdx].Length = currentOffset - openTags[i].tag.Offset
					openTags = append(openTags[:i], openTags[i+1:]...)
					matched = true
					break
				}
			}
			if !matched {
				textBuf.WriteString(token.text)
			}
		} else {
			textBuf.WriteString(token.text)
		}
	}

	currentOffset := utf16RuneCountInString(textBuf.String())
	for _, openTag := range openTags {
		tagOffsets[openTag.tagIdx].Length = currentOffset - openTag.tag.Offset
	}

	originalText := textBuf.String()
	cleanedText := strings.TrimSpace(originalText)

	leadingTrimmed := utf16RuneCountInString(originalText) -
		utf16RuneCountInString(strings.TrimLeft(originalText, " \t\n\r"))
	cleanedTextLen := utf16RuneCountInString(cleanedText)

	var newTagOffsets []tag
	for _, t := range tagOffsets {
		newOffset := max(t.Offset-leadingTrimmed, 0)
		endPos := min(t.Offset+t.Length-leadingTrimmed, cleanedTextLen)
		newLength := endPos - newOffset

		if newLength > 0 || t.Type == "emoji" || t.Type == "tg-emoji" {
			newTagOffsets = append(newTagOffsets, tag{
				Type:      t.Type,
				Length:    newLength,
				Offset:    newOffset,
				HasNested: t.HasNested,
				Attrs:     t.Attrs,
			})
		}
	}

	return cleanedText, newTagOffsets, nil
}

// supportedTag returns true if the tag is supported by Telegram.
func supportedTag(tag string) bool {
	switch tag {
	case "b", "strong", "i", "em", "u", "s", "strike", "del", "ins",
		"a", "code", "pre", "span", "spoiler", "tg-spoiler", "quote", "blockquote", "emoji", "tg-emoji",
		"mention", "br":
		return true
	}
	return false
}

// htmlUnescape unescapes HTML entities in the string.
func htmlUnescape(s string) string {
	replacer := strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
		"&#39;", "'",
		"&amp;", "&",
	)
	return replacer.Replace(s)
}

// htmlEscape escapes HTML special characters in the string.
func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}
