package parsemode

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

// FormatterMode represents the formatting style (Markdown or HTML).
type FormatterMode string

const (
	// FormatterMarkdown uses MarkdownV2 style formatting.
	FormatterMarkdown FormatterMode = "markdown"
	// FormatterHTML uses HTML style formatting.
	FormatterHTML FormatterMode = "html"
)

// FormattedText represents text with its Telegram entities.
type FormattedText struct {
	Text     string
	Entities []tg.MessageEntityClass
}

// FormatHelper provides methods to format text with Telegram entities.
type FormatHelper struct {
	mode FormatterMode
}

// NewFormatHelper creates a new format helper with the specified mode.
func NewFormatHelper(mode FormatterMode) *FormatHelper {
	return &FormatHelper{mode: mode}
}

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

func escapeHTML(text string) string {
	var result strings.Builder
	for _, r := range text {
		switch r {
		case '&':
			result.WriteString("&amp;")
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '"':
			result.WriteString("&quot;")
		default:
			result.WriteString(string(r))
		}
	}
	return result.String()
}

// Bold formats text as bold.
func (h *FormatHelper) Bold(text string) FormattedText {
	escapedText := text
	switch h.mode {
	case FormatterMarkdown:
		escapedText = escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("*%s*", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText = escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<b>%s</b>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: len(text)},
	}}
}

// Italic formats text as italic.
func (h *FormatHelper) Italic(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("_%s_", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<i>%s</i>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityItalic{Offset: 0, Length: len(text)},
	}}
}

// Underline formats text as underlined.
func (h *FormatHelper) Underline(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("__%s__", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityUnderline{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<u>%s</u>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityUnderline{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityUnderline{Offset: 0, Length: len(text)},
	}}
}

// Strikethrough formats text as strikethrough.
func (h *FormatHelper) Strikethrough(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("~%s~", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityStrike{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<s>%s</s>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityStrike{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityStrike{Offset: 0, Length: len(text)},
	}}
}

// Spoiler formats text as spoiler.
func (h *FormatHelper) Spoiler(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("||%s||", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntitySpoiler{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<span class=\"tg-spoiler\">%s</span>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntitySpoiler{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntitySpoiler{Offset: 0, Length: len(text)},
	}}
}

// Code formats text as inline code.
func (h *FormatHelper) Code(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: fmt.Sprintf("`%s`", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCode{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<code>%s</code>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCode{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityCode{Offset: 0, Length: len(text)},
	}}
}

// Pre formats text as pre-formatted code block.
func (h *FormatHelper) Pre(text string) FormattedText {
	return h.PreWithLanguage(text, "")
}

// PreWithLanguage formats text as pre-formatted code block with language.
func (h *FormatHelper) PreWithLanguage(text, language string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		if language != "" {
			return FormattedText{
				Text: fmt.Sprintf("```%s\n%s\n```", language, text),
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: language},
				},
			}
		}
		return FormattedText{
			Text: fmt.Sprintf("```\n%s\n```", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: ""},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		if language != "" {
			return FormattedText{
				Text: fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, language, escapedText),
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: language},
				},
			}
		}
		return FormattedText{
			Text: fmt.Sprintf("<pre>%s</pre>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: ""},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: language},
	}}
}

// TextLink creates an inline URL link.
func (h *FormatHelper) TextLink(text, url string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", escapedText, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<a href="%s">%s</a>`, url, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
	}}
}

// Mention creates an inline mention of a user.
func (h *FormatHelper) Mention(text string, userID int64) FormattedText {
	url := fmt.Sprintf("tg://user?id=%d", userID)
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", escapedText, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<a href="%s">%s</a>`, url, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
	}}
}

// CustomEmoji creates a custom emoji.
func (h *FormatHelper) CustomEmoji(emoji string, emojiID int64) FormattedText {
	url := fmt.Sprintf("tg://emoji?id=%d", emojiID)
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", emoji, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{Offset: 0, Length: len(emoji), DocumentID: emojiID},
			},
		}
	case FormatterHTML:
		return FormattedText{
			Text: fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, emojiID, emoji),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{Offset: 0, Length: len(emoji), DocumentID: emojiID},
			},
		}
	}
	return FormattedText{Text: emoji, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityCustomEmoji{Offset: 0, Length: len(emoji), DocumentID: emojiID},
	}}
}

// Blockquote creates a block quotation.
func (h *FormatHelper) Blockquote(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: fmt.Sprintf(">%s", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<blockquote>%s</blockquote>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text)},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: len(text)},
	}}
}

// ExpandableBlockquote creates an expandable block quotation.
func (h *FormatHelper) ExpandableBlockquote(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: fmt.Sprintf(">%s||", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text), Collapsed: true},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<blockquote expandable>%s</blockquote>`, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text), Collapsed: true},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: len(text), Collapsed: true},
	}}
}

// Combine combines multiple formatted texts into one.
// It properly adjusts the entity offsets.
func Combine(formatted ...FormattedText) FormattedText {
	if len(formatted) == 0 {
		return FormattedText{}
	}
	if len(formatted) == 1 {
		return formatted[0]
	}

	result := FormattedText{
		Text:     "",
		Entities: make([]tg.MessageEntityClass, 0),
	}

	currentOffset := 0
	for _, f := range formatted {
		result.Text += f.Text
		for _, entity := range f.Entities {
			switch e := entity.(type) {
			case *tg.MessageEntityBold:
				e.Offset += currentOffset
			case *tg.MessageEntityItalic:
				e.Offset += currentOffset
			case *tg.MessageEntityUnderline:
				e.Offset += currentOffset
			case *tg.MessageEntityStrike:
				e.Offset += currentOffset
			case *tg.MessageEntitySpoiler:
				e.Offset += currentOffset
			case *tg.MessageEntityCode:
				e.Offset += currentOffset
			case *tg.MessageEntityPre:
				e.Offset += currentOffset
			case *tg.MessageEntityTextURL:
				e.Offset += currentOffset
			case *tg.MessageEntityCustomEmoji:
				e.Offset += currentOffset
			case *tg.MessageEntityBlockquote:
				e.Offset += currentOffset
			}
			result.Entities = append(result.Entities, entity)
		}
		currentOffset += len(f.Text)
	}

	return result
}
