package parsemode

import "fmt"

// AddBold wraps text with bold formatting markers based on formatter mode.
func (h *FormatHelper) AddBold(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("*%s*", text)
	case FormatterHTML:
		return fmt.Sprintf("<b>%s</b>", htmlEscape(text))
	default:
		return text
	}
}

// AddItalic wraps text with italic formatting markers.
func (h *FormatHelper) AddItalic(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("_%s_", text)
	case FormatterHTML:
		return fmt.Sprintf("<i>%s</i>", htmlEscape(text))
	default:
		return text
	}
}

// AddUnderline wraps text with underline formatting markers.
func (h *FormatHelper) AddUnderline(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("__%s__", text)
	case FormatterHTML:
		return fmt.Sprintf("<u>%s</u>", htmlEscape(text))
	default:
		return text
	}
}

// AddStrikethrough wraps text with strikethrough formatting markers.
func (h *FormatHelper) AddStrikethrough(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("~%s~", text)
	case FormatterHTML:
		return fmt.Sprintf("<s>%s</s>", htmlEscape(text))
	default:
		return text
	}
}

// AddSpoiler wraps text with spoiler formatting markers.
func (h *FormatHelper) AddSpoiler(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("||%s||", text)
	case FormatterHTML:
		return fmt.Sprintf("<tg-spoiler>%s</tg-spoiler>", htmlEscape(text))
	default:
		return text
	}
}

// AddCode wraps text with code formatting markers.
func (h *FormatHelper) AddCode(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("`%s`", text)
	case FormatterHTML:
		return fmt.Sprintf("<code>%s</code>", htmlEscape(text))
	default:
		return text
	}
}

// AddPre wraps text with pre-formatted code block markers.
func (h *FormatHelper) AddPre(text, language string) string {
	switch h.mode {
	case FormatterMarkdown:
		if language != "" {
			return fmt.Sprintf("```%s\n%s\n```", language, text)
		}
		return fmt.Sprintf("```\n%s\n```", text)
	case FormatterHTML:
		if language != "" {
			return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", htmlEscape(language), htmlEscape(text))
		}
		return fmt.Sprintf("<pre>%s</pre>", htmlEscape(text))
	default:
		return text
	}
}

// CreateEmbedLink creates a clickable link/mention for Telegram.
// For user mentions: link should be "tg://user?id=USERID"
// For URLs: link should be "https://example.com"
func (h *FormatHelper) CreateEmbedLink(text, link string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("[%s](%s)", text, link)
	case FormatterHTML:
		return fmt.Sprintf(`<a href="%s">%s</a>`, htmlEscape(link), htmlEscape(text))
	default:
		return text
	}
}

// CreateUserMention creates a mention link for a Telegram user.
// The displayName is the visible text, userID is the user's Telegram ID.
func (h *FormatHelper) CreateUserMention(displayName string, userID int64) string {
	link := fmt.Sprintf("tg://user?id=%d", userID)
	return h.CreateEmbedLink(displayName, link)
}

// CreateLink creates a hyperlink with the specified URL.
func (h *FormatHelper) CreateLink(text, url string) string {
	return h.CreateEmbedLink(text, url)
}

// CreateCustomEmoji creates a custom emoji link.
// emoji is the emoji character, emojiID is the custom emoji ID from Telegram.
func (h *FormatHelper) CreateCustomEmoji(emoji string, emojiID int64) string {
	link := fmt.Sprintf("tg://emoji?id=%d", emojiID)
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("[%s](%s)", emoji, link)
	case FormatterHTML:
		return fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, emojiID, htmlEscape(emoji))
	default:
		return emoji
	}
}

// AddBlockquote wraps text with blockquote formatting markers.
func (h *FormatHelper) AddBlockquote(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf(">%s", text)
	case FormatterHTML:
		return fmt.Sprintf("<blockquote>%s</blockquote>", htmlEscape(text))
	default:
		return text
	}
}

// AddExpandableBlockquote wraps text with expandable blockquote formatting markers.
func (h *FormatHelper) AddExpandableBlockquote(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf(">%s||", text)
	case FormatterHTML:
		return fmt.Sprintf("<blockquote expandable>%s</blockquote>", htmlEscape(text))
	default:
		return text
	}
}
