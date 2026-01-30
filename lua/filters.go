package lua

import (
	"slices"
	"strings"

	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/types"
)

var commandFilter = func(m *types.Message) bool {
	if m.Text == "" {
		return false
	}
	return slices.Contains(handlers.DefaultPrefix, rune(m.Text[0]))
}

func resolveUpdateFilter(name string) filters.UpdateFilter {
	switch name {
	case "private":
		return filters.Private
	case "group":
		return filters.Group
	case "supergroup":
		return filters.Supergroup
	case "channel":
		return filters.Channel
	case "incoming":
		return filters.Incoming
	case "outgoing":
		return filters.Outgoing
	case "business":
		return filters.Business
	default:
		return nil
	}
}

func resolveMessageFilter(name string) filters.MessageFilter {
	if strings.HasPrefix(name, "!") {
		inner := resolveMessageFilter(name[1:])
		if inner == nil {
			return nil
		}
		return filters.MessageNot(inner)
	}

	switch name {
	case "command":
		return commandFilter
	case "text":
		return filters.Message.Text
	case "photo":
		return filters.Message.Photo
	case "video":
		return filters.Message.Video
	case "audio":
		return filters.Message.Audio
	case "voice":
		return filters.Message.Voice
	case "sticker":
		return filters.Message.Sticker
	case "animation":
		return filters.Message.Animation
	case "document":
		return filters.Document
	case "media":
		return filters.Message.Media
	case "reply":
		return filters.Message.Reply
	case "forwarded":
		return filters.Message.Forwarded
	case "edited":
		return filters.Message.Edited
	case "poll":
		return filters.Message.Poll
	case "dice":
		return filters.Message.Dice
	case "game":
		return filters.Message.Game
	case "contact":
		return filters.Message.Contact
	case "location":
		return filters.Message.Location
	case "venue":
		return filters.Message.Venue
	case "video_note":
		return filters.Message.VideoNote
	case "web_page":
		return filters.Message.WebPage
	case "private":
		return filters.Message.Private
	case "group":
		return filters.Message.Group
	case "channel":
		return filters.Message.Channel
	case "incoming":
		return filters.Message.Incoming
	case "outgoing":
		return filters.Message.Outgoing
	default:
		return nil
	}
}
