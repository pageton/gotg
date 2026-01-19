package parsemode

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
)

// HTMLParser implements the Parser interface for HTML formatting.
// It is thread-safe and can be used concurrently.
type HTMLParser struct {
	// builderPool allows reusing entity builders for better performance.
	// Using nil for now as Go's sync.Pool might be overkill for this simple use case.
}

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
		// On error, return the original text with no entities
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
			// Find the end of the tag
			tagEnd := i + 1
			for tagEnd < len(html) && html[tagEnd] != '>' {
				tagEnd++
			}

			if tagEnd >= len(html) {
				// Unclosed tag, treat as text
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

				// Parse attributes
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
				// Empty tag, treat as text
				tokens = append(tokens, htmlToken{
					isTag: false,
					text:  html[i : tagEnd+1],
				})
			}
			i = tagEnd + 1
		} else {
			// Regular text
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

	// Stack for tracking open tags
	type openTag struct {
		tag    tag
		tagIdx int
	}
	var openTags []openTag

	for _, token := range tokens {
		if !token.isTag {
			// Text content - unescape HTML entities
			textBuf.WriteString(htmlUnescape(token.text))
		} else if !token.isClosing && supportedTag(token.tagName) {
			// Opening tag
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
			// Closing tag - find matching opening tag
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
				// No matching opening tag, treat as text
				textBuf.WriteString(token.text)
			}
		} else {
			// Unsupported tag, treat as text
			textBuf.WriteString(token.text)
		}
	}

	// Close any unclosed tags
	currentOffset := utf16RuneCountInString(textBuf.String())
	for _, openTag := range openTags {
		tagOffsets[openTag.tagIdx].Length = currentOffset - openTag.tag.Offset
	}

	// Trim whitespace and adjust offsets
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

		// Allow zero-length entities for custom emoji (Telegram renders the actual emoji)
		if newLength > 0 || t.Type == "emoji" {
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
		"a", "code", "pre", "spoiler", "tg-spoiler", "quote", "blockquote", "emoji",
		"mention", "br":
		return true
	}
	return false
}

// htmlUnescape unescapes HTML entities in the string.
func htmlUnescape(s string) string {
	// Order matters - & must be last
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

// parseTagsToEntities converts parsed tags into Telegram message entities.
func (p *HTMLParser) parseTagsToEntities(tags []tag) []tg.MessageEntityClass {
	if len(tags) == 0 {
		return nil
	}

	entities := make([]tg.MessageEntityClass, 0, len(tags))

	for _, tag := range tags {
		entity := p.tagToEntity(tag)
		if entity != nil {
			entities = append(entities, entity)
		}
	}

	return entities
}

// tagToEntity converts a single tag to a Telegram message entity.
func (p *HTMLParser) tagToEntity(tag tag) tg.MessageEntityClass {
	switch tag.Type {
	case "a":
		return p.parseAnchorTag(tag)
	case "b", "strong":
		return &tg.MessageEntityBold{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "code":
		return &tg.MessageEntityCode{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "em", "i":
		return &tg.MessageEntityItalic{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "pre":
		language := tag.Attrs["class"]
		if after, ok := strings.CutPrefix(language, "language-"); ok {
			language = after
		}
		return &tg.MessageEntityPre{
			Offset:   int(tag.Offset),
			Length:   int(tag.Length),
			Language: language,
		}
	case "s", "strike", "del":
		return &tg.MessageEntityStrike{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "u", "ins":
		return &tg.MessageEntityUnderline{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "spoiler", "tg-spoiler":
		return &tg.MessageEntitySpoiler{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case "quote", "blockquote":
		collapsed := false
		if c, ok := tag.Attrs["collapsed"]; ok {
			collapsed, _ = strconv.ParseBool(c)
		}
		if _, hasExpandable := tag.Attrs["expandable"]; hasExpandable {
			collapsed = true
		}
		return &tg.MessageEntityBlockquote{
			Collapsed: collapsed,
			Offset:    int(tag.Offset),
			Length:    int(tag.Length),
		}
	case "emoji":
		emojiID, err := strconv.ParseInt(tag.Attrs["id"], 10, 64)
		if err != nil {
			return nil
		}
		return &tg.MessageEntityCustomEmoji{
			Offset:     int(tag.Offset),
			Length:     int(tag.Length),
			DocumentID: emojiID,
		}
	case "mention":
		return &tg.MessageEntityMention{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	}

	return nil
}

// parseAnchorTag converts an <a> tag to the appropriate Telegram entity.
func (p *HTMLParser) parseAnchorTag(tag tag) tg.MessageEntityClass {
	href := tag.Attrs["href"]

	switch {
	case href == "":
		// Link without href is treated as plain URL
		return &tg.MessageEntityURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case strings.HasPrefix(href, "mailto:"):
		// Email link
		return &tg.MessageEntityEmail{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case strings.HasPrefix(href, "tg://emoji?id="):
		// Custom emoji link
		u, err := url.Parse(href)
		if err == nil {
			id := u.Query().Get("id")
			if id != "" {
				emojiID, err := strconv.ParseInt(id, 10, 64)
				if err == nil {
					return &tg.MessageEntityCustomEmoji{
						Offset:     int(tag.Offset),
						Length:     int(tag.Length),
						DocumentID: emojiID,
					}
				}
			}
		}
		return &tg.MessageEntityTextURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
			URL:    href,
		}
	case strings.HasPrefix(href, "tg://user?id="):
		// User mention link
		idStr := strings.TrimPrefix(href, "tg://user?id=")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			return &tg.InputMessageEntityMentionName{
				Offset: int(tag.Offset),
				Length: int(tag.Length),
				UserID: &tg.InputUser{
					UserID:     id,
					AccessHash: 0,
				},
			}
		}
		return &tg.MessageEntityTextURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
			URL:    href,
		}
	default:
		// Regular link with URL
		return &tg.MessageEntityTextURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
			URL:    href,
		}
	}
}

// Format formats text as HTML (no-op for HTML parser, returns input as-is).
func (p *HTMLParser) Format(input string) string {
	return input
}

// FormatEntity formats a single entity as HTML.
func FormatEntity(entityType string, content string, attrs map[string]string) string {
	switch entityType {
	case string(EntityTypeBold):
		return fmt.Sprintf("<b>%s</b>", content)
	case string(EntityTypeItalic):
		return fmt.Sprintf("<i>%s</i>", content)
	case string(EntityTypeUnderline):
		return fmt.Sprintf("<u>%s</u>", content)
	case string(EntityTypeStrike):
		return fmt.Sprintf("<s>%s</s>", content)
	case string(EntityTypeSpoiler):
		return fmt.Sprintf("<spoiler>%s</spoiler>", content)
	case string(EntityTypeCode):
		return fmt.Sprintf("<code>%s</code>", content)
	case string(EntityTypePre):
		lang := ""
		if l, ok := attrs["language"]; ok && l != "" {
			lang = fmt.Sprintf(` class="language-%s"`, l)
		}
		return fmt.Sprintf("<pre%s>%s</pre>", lang, content)
	case string(EntityTypeBlockquote):
		collapsed := ""
		if c, ok := attrs["collapsed"]; ok && c == "true" {
			collapsed = ` collapsed="true"`
		}
		return fmt.Sprintf("<blockquote%s>%s</blockquote>", collapsed, content)
	case string(EntityTypeTextURL):
		url := attrs["url"]
		if url == "" {
			url = content
		}
		return fmt.Sprintf(`<a href="%s">%s</a>`, htmlEscape(url), content)
	case string(EntityTypeMention):
		return fmt.Sprintf(`<mention>%s</mention>`, content)
	case string(EntityTypeCustomEmoji):
		emojiID := attrs["emoji_id"]
		return fmt.Sprintf(`<emoji id="%s">%s</emoji>`, emojiID, content)
	default:
		return content
	}
}
