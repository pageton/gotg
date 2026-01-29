package adapter

import (
	"strconv"

	"github.com/pageton/gotg/parsemode"
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
		return "<b>" + text + "</b>"
	case Markdown, MarkdownV2:
		return "*" + escapeMarkdownV2(text) + "*"
	}
	return text
}

// Italic wraps text with italic formatting markers.
func (f *FormatHelper) Italic(text string) string {
	switch f.mode {
	case HTML:
		return "<i>" + text + "</i>"
	case Markdown, MarkdownV2:
		return "_" + escapeMarkdownV2(text) + "_"
	}
	return text
}

// Underline wraps text with underline formatting markers.
func (f *FormatHelper) Underline(text string) string {
	switch f.mode {
	case HTML:
		return "<u>" + text + "</u>"
	case Markdown, MarkdownV2:
		return "__" + escapeMarkdownV2(text) + "__"
	}
	return text
}

// Strikethrough wraps text with strikethrough formatting markers.
func (f *FormatHelper) Strikethrough(text string) string {
	switch f.mode {
	case HTML:
		return "<s>" + text + "</s>"
	case Markdown, MarkdownV2:
		return "~" + escapeMarkdownV2(text) + "~"
	}
	return text
}

// Spoiler wraps text with spoiler formatting markers.
func (f *FormatHelper) Spoiler(text string) string {
	switch f.mode {
	case HTML:
		return "<tg-spoiler>" + text + "</tg-spoiler>"
	case Markdown, MarkdownV2:
		return "||" + escapeMarkdownV2(text) + "||"
	}
	return text
}

// Code wraps text with code formatting markers.
func (f *FormatHelper) Code(text string) string {
	switch f.mode {
	case HTML:
		return "<code>" + text + "</code>"
	case Markdown, MarkdownV2:
		return "`" + text + "`"
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
			return `<pre><code class="language-` + language + `">` + text + "</code></pre>"
		}
		return "<pre>" + text + "</pre>"
	case Markdown, MarkdownV2:
		if language != "" {
			return "```" + language + "\n" + text + "\n```"
		}
		return "```\n" + text + "\n```"
	}
	return text
}

// Link creates a hyperlink with the specified URL.
func (f *FormatHelper) Link(text, url string) string {
	switch f.mode {
	case HTML:
		return "<a href='" + url + "'>" + text + "</a>"
	case Markdown, MarkdownV2:
		return "[" + escapeMarkdownV2(text) + "](" + url + ")"
	}
	return text + ": " + url
}

// Mention creates a mention link for a Telegram user.
func (f *FormatHelper) Mention(displayName string, userID int64) string {
	link := "tg://user?id=" + strconv.FormatInt(userID, 10)
	return f.Link(displayName, link)
}

// CustomEmoji creates a custom emoji link.
func (f *FormatHelper) CustomEmoji(emoji string, emojiID int64) string {
	idStr := strconv.FormatInt(emojiID, 10)
	switch f.mode {
	case HTML:
		return `<tg-emoji emoji-id="` + idStr + `">` + emoji + "</tg-emoji>"
	case Markdown, MarkdownV2:
		link := "tg://emoji?id=" + idStr
		return "[" + emoji + "](" + link + ")"
	}
	return emoji
}

// Blockquote wraps text with blockquote formatting markers.
func (f *FormatHelper) Blockquote(text string) string {
	switch f.mode {
	case HTML:
		return "<blockquote>" + text + "</blockquote>"
	case Markdown, MarkdownV2:
		return ">" + text
	}
	return text
}

// ExpandableBlockquote wraps text with expandable blockquote formatting markers.
func (f *FormatHelper) ExpandableBlockquote(text string) string {
	switch f.mode {
	case HTML:
		return "<blockquote expandable>" + text + "</blockquote>"
	case Markdown, MarkdownV2:
		return ">" + text + "||"
	}
	return text
}

var escapeMarkdownV2 = parsemode.EscapeMarkdownV2
