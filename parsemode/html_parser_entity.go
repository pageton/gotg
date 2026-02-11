package parsemode

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
)

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
	case "span":
		if tag.Attrs["class"] == "tg-spoiler" {
			return &tg.MessageEntitySpoiler{
				Offset: int(tag.Offset),
				Length: int(tag.Length),
			}
		}
		return nil
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
	case "tg-emoji":
		emojiID, err := strconv.ParseInt(tag.Attrs["emoji-id"], 10, 64)
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

	if !isValidTelegramURL(href) {
		return &tg.MessageEntityURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	}

	href = normalizeURL(href)

	switch {
	case href == "":
		return &tg.MessageEntityURL{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case strings.HasPrefix(href, "mailto:"):
		return &tg.MessageEntityEmail{
			Offset: int(tag.Offset),
			Length: int(tag.Length),
		}
	case strings.HasPrefix(href, "tg://emoji?id="):
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
