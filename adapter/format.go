package adapter

import (
	"fmt"
	"strings"
)

// FormatHelper provides text formatting convenience methods.
// Use Update.Format() to get a formatter instance.
type FormatHelper struct {
	mode string // HTML, Markdown, or ""
}

// Format returns a new FormatHelper for formatting text.
// The mode determines the formatting style: "HTML", "Markdown", "MarkdownV2", or "" (none).
func (u *Update) Format(mode string) *FormatHelper {
	return &FormatHelper{mode: mode}
}

// Bold wraps text with bold formatting markers.
func (f *FormatHelper) Bold(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<b>%s</b>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("*%s*", escapeMarkdownV2(text))
	}
	return text
}

// Italic wraps text with italic formatting markers.
func (f *FormatHelper) Italic(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<i>%s</i>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("_%s_", escapeMarkdownV2(text))
	}
	return text
}

// Underline wraps text with underline formatting markers.
func (f *FormatHelper) Underline(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<u>%s</u>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("__%s__", escapeMarkdownV2(text))
	}
	return text
}

// Strikethrough wraps text with strikethrough formatting markers.
func (f *FormatHelper) Strikethrough(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<s>%s</s>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("~%s~", escapeMarkdownV2(text))
	}
	return text
}

// Spoiler wraps text with spoiler formatting markers.
func (f *FormatHelper) Spoiler(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<tg-spoiler>%s</tg-spoiler>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("||%s||", escapeMarkdownV2(text))
	}
	return text
}

// Code wraps text with code formatting markers.
func (f *FormatHelper) Code(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<code>%s</code>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("`%s`", text)
	}
	return text
}

// Pre wraps text with pre-formatted code block markers.
func (f *FormatHelper) Pre(text string) string {
	return f.PreWithLanguage(text, "")
}

// PreWithLanguage wraps text with pre-formatted code block markers with syntax highlighting.
func (f *FormatHelper) PreWithLanguage(text, language string) string {
	switch f.mode {
	case HTML:
		if language != "" {
			return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", language, text)
		}
		return fmt.Sprintf("<pre>%s</pre>", text)
	case Markdown, MarkdownV2:
		if language != "" {
			return fmt.Sprintf("```%s\n%s\n```", language, text)
		}
		return fmt.Sprintf("```\n%s\n```", text)
	}
	return text
}

// Link creates a hyperlink with the specified URL.
func (f *FormatHelper) Link(text, url string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<a href='%s'>%s</a>", url, text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("[%s](%s)", escapeMarkdownV2(text), url)
	}
	return fmt.Sprintf("%s: %s", text, url)
}

// Mention creates a mention link for a Telegram user.
func (f *FormatHelper) Mention(displayName string, userID int64) string {
	link := fmt.Sprintf("tg://user?id=%d", userID)
	return f.Link(displayName, link)
}

// CustomEmoji creates a custom emoji link.
func (f *FormatHelper) CustomEmoji(emoji string, emojiID int64) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<tg-emoji emoji-id=\"%d\">%s</tg-emoji>", emojiID, emoji)
	case Markdown, MarkdownV2:
		link := fmt.Sprintf("tg://emoji?id=%d", emojiID)
		return fmt.Sprintf("[%s](%s)", emoji, link)
	}
	return emoji
}

// Blockquote wraps text with blockquote formatting markers.
func (f *FormatHelper) Blockquote(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<blockquote>%s</blockquote>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf(">%s", text)
	}
	return text
}

// ExpandableBlockquote wraps text with expandable blockquote formatting markers.
func (f *FormatHelper) ExpandableBlockquote(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<blockquote expandable>%s</blockquote>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf(">%s||", text)
	}
	return text
}

// escapeMarkdownV2 escapes special characters for MarkdownV2.
func escapeMarkdownV2(text string) string {
	escapeMap := map[rune]string{
		'_': "\\_",
		'*': "\\*",
		'[': "\\[",
		']': "\\]",
		'(': "\\(",
		')': "\\)",
		'~': "\\~",
		'`': "\\`",
		'>': "\\>",
		'#': "\\#",
		'+': "\\+",
		'-': "\\-",
		'=': "\\=",
		'|': "\\|",
		'{': "\\{",
		'}': "\\}",
		'.': "\\.",
		'!': "\\!",
	}
	var result strings.Builder
	for _, r := range text {
		if escaped, ok := escapeMap[r]; ok {
			result.WriteString(escaped)
		} else {
			result.WriteString(string(r))
		}
	}
	return result.String()
}
