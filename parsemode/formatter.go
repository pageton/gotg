package parsemode

import (
	"strconv"
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

var (
	escapeMarkdownV2 = EscapeMarkdownV2
	escapeHTML       = EscapeHTML
)

// Bold formats text as bold.
func (h *FormatHelper) Bold(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: "*" + escapedText + "*",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<b>" + escapedText + "</b>",
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
			Text: "_" + escapedText + "_",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<i>" + escapedText + "</i>",
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
			Text: "__" + escapedText + "__",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityUnderline{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<u>" + escapedText + "</u>",
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
			Text: "~" + escapedText + "~",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityStrike{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<s>" + escapedText + "</s>",
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
			Text: "||" + escapedText + "||",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntitySpoiler{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: `<span class="tg-spoiler">` + escapedText + "</span>",
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
			Text: "`" + text + "`",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCode{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<code>" + escapedText + "</code>",
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
				Text: "```" + language + "\n" + text + "\n```",
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: language},
				},
			}
		}
		return FormattedText{
			Text: "```\n" + text + "\n```",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: ""},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		if language != "" {
			return FormattedText{
				Text: `<pre><code class="language-` + language + `">` + escapedText + "</code></pre>",
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{Offset: 0, Length: len(text), Language: language},
				},
			}
		}
		return FormattedText{
			Text: "<pre>" + escapedText + "</pre>",
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
			Text: "[" + escapedText + "](" + url + ")",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: `<a href="` + url + `">` + escapedText + "</a>",
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
	url := "tg://user?id=" + strconv.FormatInt(userID, 10)
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: "[" + escapedText + "](" + url + ")",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 0, Length: len(text), URL: url},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: `<a href="` + url + `">` + escapedText + "</a>",
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
	idStr := strconv.FormatInt(emojiID, 10)
	url := "tg://emoji?id=" + idStr
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: "[" + emoji + "](" + url + ")",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{Offset: 0, Length: len(emoji), DocumentID: emojiID},
			},
		}
	case FormatterHTML:
		return FormattedText{
			Text: `<tg-emoji emoji-id="` + idStr + `">` + emoji + "</tg-emoji>",
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
			Text: ">" + text,
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text)},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<blockquote>" + escapedText + "</blockquote>",
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
			Text: ">" + text + "||",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: len(text), Collapsed: true},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: "<blockquote expandable>" + escapedText + "</blockquote>",
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

	totalLen := 0
	totalEntities := 0
	for _, f := range formatted {
		totalLen += len(f.Text)
		totalEntities += len(f.Entities)
	}

	var sb strings.Builder
	sb.Grow(totalLen)
	allEntities := make([]tg.MessageEntityClass, 0, totalEntities)

	currentOffset := 0
	for _, f := range formatted {
		sb.WriteString(f.Text)
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
			allEntities = append(allEntities, entity)
		}
		currentOffset += len(f.Text)
	}

	return FormattedText{
		Text:     sb.String(),
		Entities: allEntities,
	}
}
