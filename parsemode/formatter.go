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
	// Escape special characters in MarkdownV2
	// Order matters for some characters
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
// Markdown: *bold*
// HTML: <b>bold</b>
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
// Markdown: _italic_
// HTML: <i>italic</i>
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
// Markdown: __underline__
// HTML: <u>underline</u>
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
// Markdown: ~strikethrough~
// HTML: <s>strikethrough</s>
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
// Markdown: ||spoiler||
// HTML: <span class="tg-spoiler">spoiler</span>
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
// Markdown: `code`
// HTML: <code>code</code>
func (h *FormatHelper) Code(text string) FormattedText {
	// Code doesn't need escaping in Markdown, but we should handle backticks
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
// Markdown: ```code```
// HTML: <pre>code</pre>
func (h *FormatHelper) Pre(text string) FormattedText {
	return h.PreWithLanguage(text, "")
}

// PreWithLanguage formats text as pre-formatted code block with language.
// Markdown: ```python\ncode\n```
// HTML: <pre><code class="language-python">code</code></pre>
func (h *FormatHelper) PreWithLanguage(text, language string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		if language != "" {
			return FormattedText{
				Text: fmt.Sprintf("```%s\n%s\n```", language, text),
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{
						Offset:   0,
						Length:   len(text),
						Language: language,
					},
				},
			}
		}
		return FormattedText{
			Text: fmt.Sprintf("```\n%s\n```", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{
					Offset:   0,
					Length:   len(text),
					Language: "",
				},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		if language != "" {
			return FormattedText{
				Text: fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, language, escapedText),
				Entities: []tg.MessageEntityClass{
					&tg.MessageEntityPre{
						Offset:   0,
						Length:   len(text),
						Language: language,
					},
				},
			}
		}
		return FormattedText{
			Text: fmt.Sprintf("<pre>%s</pre>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{
					Offset:   0,
					Length:   len(text),
					Language: "",
				},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityPre{
			Offset:   0,
			Length:   len(text),
			Language: language,
		},
	}}
}

// TextLink creates an inline URL link.
// Markdown: [text](url)
// HTML: <a href="url">text</a>
func (h *FormatHelper) TextLink(text, url string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", escapedText, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{
					Offset: 0,
					Length: len(text),
					URL:    url,
				},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<a href="%s">%s</a>`, url, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{
					Offset: 0,
					Length: len(text),
					URL:    url,
				},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{
			Offset: 0,
			Length: len(text),
			URL:    url,
		},
	}}
}

// Mention creates an inline mention of a user.
// Markdown: [text](tg://user?id=123456789)
// HTML: <a href="tg://user?id=123456789">text</a>
func (h *FormatHelper) Mention(text string, userID int64) FormattedText {
	url := fmt.Sprintf("tg://user?id=%d", userID)
	switch h.mode {
	case FormatterMarkdown:
		escapedText := escapeMarkdownV2(text)
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", escapedText, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{
					Offset: 0,
					Length: len(text),
					URL:    url,
				},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<a href="%s">%s</a>`, url, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{
					Offset: 0,
					Length: len(text),
					URL:    url,
				},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{
			Offset: 0,
			Length: len(text),
			URL:    url,
		},
	}}
}

// CustomEmoji creates a custom emoji.
// Markdown: ![emoji](tg://emoji?id=123456789)
// HTML: <tg-emoji emoji-id="123456789">emoji</tg-emoji>
func (h *FormatHelper) CustomEmoji(emoji string, emojiID int64) FormattedText {
	url := fmt.Sprintf("tg://emoji?id=%d", emojiID)
	switch h.mode {
	case FormatterMarkdown:
		return FormattedText{
			Text: fmt.Sprintf("[%s](%s)", emoji, url),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{
					Offset:     0,
					Length:     len(emoji),
					DocumentID: emojiID,
				},
			},
		}
	case FormatterHTML:
		return FormattedText{
			Text: fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, emojiID, emoji),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{
					Offset:     0,
					Length:     len(emoji),
					DocumentID: emojiID,
				},
			},
		}
	}
	return FormattedText{Text: emoji, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityCustomEmoji{
			Offset:     0,
			Length:     len(emoji),
			DocumentID: emojiID,
		},
	}}
}

// Blockquote creates a block quotation.
// Markdown: >text
// HTML: <blockquote>text</blockquote>
func (h *FormatHelper) Blockquote(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		// Blockquote in MarkdownV2: each line starts with >
		return FormattedText{
			Text: fmt.Sprintf(">%s", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{
					Offset: 0,
					Length: len(text),
				},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf("<blockquote>%s</blockquote>", escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{
					Offset: 0,
					Length: len(text),
				},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{
			Offset: 0,
			Length: len(text),
		},
	}}
}

// ExpandableBlockquote creates an expandable block quotation.
// Markdown: >text (with expandable mark ||)
// HTML: <blockquote expandable>text</blockquote>
func (h *FormatHelper) ExpandableBlockquote(text string) FormattedText {
	switch h.mode {
	case FormatterMarkdown:
		// Expandable blockquote in MarkdownV2
		return FormattedText{
			Text: fmt.Sprintf(">%s||", text),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{
					Offset:    0,
					Length:    len(text),
					Collapsed: true,
				},
			},
		}
	case FormatterHTML:
		escapedText := escapeHTML(text)
		return FormattedText{
			Text: fmt.Sprintf(`<blockquote expandable>%s</blockquote>`, escapedText),
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{
					Offset:    0,
					Length:    len(text),
					Collapsed: true,
				},
			},
		}
	}
	return FormattedText{Text: text, Entities: []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{
			Offset:    0,
			Length:    len(text),
			Collapsed: true,
		},
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
			// Adjust offset for each entity based on current position
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

// AddBold wraps text with bold formatting markers based on formatter mode.
func (h *FormatHelper) AddBold(text string) string {
	switch h.mode {
	case FormatterMarkdown:
		return fmt.Sprintf("*%s*", text)
	case FormatterHTML:
		return fmt.Sprintf("<b>%s</b>", text)
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
		return fmt.Sprintf("<i>%s</i>", text)
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
		return fmt.Sprintf("<u>%s</u>", text)
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
		return fmt.Sprintf("<s>%s</s>", text)
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
		return fmt.Sprintf("<tg-spoiler>%s</tg-spoiler>", text)
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
		return fmt.Sprintf("<code>%s</code>", text)
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
			return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", language, text)
		}
		return fmt.Sprintf("<pre>%s</pre>", text)
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
		return fmt.Sprintf("<a href='%s'>%s</a>", link, text)
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
		return fmt.Sprintf("<tg-emoji emoji-id=\"%d\">%s</tg-emoji>", emojiID, emoji)
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
		return fmt.Sprintf("<blockquote>%s</blockquote>", text)
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
		return fmt.Sprintf("<blockquote expandable>%s</blockquote>", text)
	default:
		return text
	}
}
