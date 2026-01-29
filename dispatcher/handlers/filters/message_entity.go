package filters

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

// Mention returns true if the Message contains a mention.
func (*messageFilters) Mention(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityMention); ok {
			return true
		}
	}
	return false
}

// Hashtag returns true if the Message contains a hashtag.
func (*messageFilters) Hashtag(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityHashtag); ok {
			return true
		}
	}
	return false
}

// Cashtag returns true if the Message contains a cashtag.
func (*messageFilters) Cashtag(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityCashtag); ok {
			return true
		}
	}
	return false
}

// BotCommand returns true if the Message contains a bot command.
func (*messageFilters) BotCommand(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityBotCommand); ok {
			return true
		}
	}
	return false
}

// Url returns true if the Message contains a URL.
func (*messageFilters) Url(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityURL); ok {
			return true
		}
		if _, ok := entity.(*tg.MessageEntityTextURL); ok {
			return true
		}
	}
	return false
}

// Email returns true if the Message contains an email.
func (*messageFilters) Email(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityEmail); ok {
			return true
		}
	}
	return false
}

// PhoneNumber returns true if the Message contains a phone number.
func (*messageFilters) PhoneNumber(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityPhone); ok {
			return true
		}
	}
	return false
}

// Bold returns true if the Message contains bold formatting.
func (*messageFilters) Bold(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityBold); ok {
			return true
		}
	}
	return false
}

// Italic returns true if the Message contains italic formatting.
func (*messageFilters) Italic(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityItalic); ok {
			return true
		}
	}
	return false
}

// Underline returns true if the Message contains underline formatting.
func (*messageFilters) Underline(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityUnderline); ok {
			return true
		}
	}
	return false
}

// Strike returns true if the Message contains strikethrough formatting.
func (*messageFilters) Strike(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityStrike); ok {
			return true
		}
	}
	return false
}

// Code returns true if the Message contains code formatting.
func (*messageFilters) Code(m *types.Message) bool {
	for _, entity := range m.Entities {
		if _, ok := entity.(*tg.MessageEntityCode); ok {
			return true
		}
	}
	return false
}

// Pre returns true if the Message contains pre-formatted text.
func (*messageFilters) Pre(m *types.Message) bool {
	for _, entity := range m.Entities {
		if pre, ok := entity.(*tg.MessageEntityPre); ok && pre != nil {
			return true
		}
	}
	return false
}

// Spoiler returns true if the Message contains spoiler formatting.
func (*messageFilters) Spoiler(m *types.Message) bool {
	for _, entity := range m.Entities {
		if spoiler, ok := entity.(*tg.MessageEntitySpoiler); ok && spoiler != nil {
			return true
		}
	}
	return false
}

// Blockquote returns true if the Message contains blockquote formatting.
func (*messageFilters) Blockquote(m *types.Message) bool {
	for _, entity := range m.Entities {
		if quote, ok := entity.(*tg.MessageEntityBlockquote); ok && quote != nil {
			return true
		}
	}
	return false
}

// CustomEmoji returns true if the Message contains a custom emoji.
func (*messageFilters) CustomEmoji(m *types.Message) bool {
	for _, entity := range m.Entities {
		if emoji, ok := entity.(*tg.MessageEntityCustomEmoji); ok && emoji != nil {
			return true
		}
	}
	return false
}
