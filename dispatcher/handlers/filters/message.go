package filters

import (
	"regexp"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/types"
)

type messageFilters struct{}

// All returns true on every type of types.Message update.
func (*messageFilters) All(_ *types.Message) bool {
	return true
}

type ChatType int

const (
	ChatTypeUser ChatType = iota
	ChatTypeChat
	ChatTypeChannel
)

func (*messageFilters) ChatType(chatType ChatType) MessageFilter {
	return func(m *types.Message) bool {
		chatPeer := m.PeerID
		switch chatType {
		case ChatTypeUser:
			_, ok := chatPeer.(*tg.PeerUser)
			return ok
		case ChatTypeChat:
			_, ok := chatPeer.(*tg.PeerChat)
			return ok
		case ChatTypeChannel:
			_, ok := chatPeer.(*tg.PeerChannel)
			return ok
		}
		return false
	}
}

// Chat allows the types.Message update to process if it is from that particular chat.
func (*messageFilters) Chat(chatID int64) MessageFilter {
	return func(m *types.Message) bool {
		return functions.GetChatIDFromPeer(m.PeerID) == chatID
	}
}

// Text returns true if types.Message consists of text.
func (*messageFilters) Text(m *types.Message) bool {
	return m.Text != ""
}

// Equal checks if the message text is equal to the provided text and returns true if matches.
func (*messageFilters) Equal(text string) MessageFilter {
	return func(m *types.Message) bool {
		return m.Text == text
	}
}

// Prefix returns true if the message text starts with the provided prefix.
func (*messageFilters) Prefix(prefix string) MessageFilter {
	return func(m *types.Message) bool {
		return strings.HasPrefix(m.Text, prefix)
	}
}

// Suffix returns true if the message text ends with the provided suffix.
func (*messageFilters) Suffix(suffix string) MessageFilter {
	return func(m *types.Message) bool {
		return strings.HasSuffix(m.Text, suffix)
	}
}

// Contains returns true if the message text contains the provided substring.
func (*messageFilters) Contains(substring string) MessageFilter {
	return func(m *types.Message) bool {
		return strings.Contains(m.Text, substring)
	}
}

// Regex returns true if the Message field of types.Message matches the regex filter
func (*messageFilters) Regex(rString string) MessageFilter {
	r := regexp.MustCompile(rString)
	return func(m *types.Message) bool {
		return r.MatchString(m.Text)
	}
}

// Media returns true if types.Message consists of media.
func (*messageFilters) Media(m *types.Message) bool {
	return m.Media != nil
}

// Photo returns true if types.Message consists of photo.
func (*messageFilters) Photo(m *types.Message) bool {
	_, photo := m.Media.(*tg.MessageMediaPhoto)
	return photo
}

// Video returns true if types.Message consists of video, gif etc.
func (*messageFilters) Video(m *types.Message) bool {
	doc := GetDocument(m)
	if doc != nil {
		for _, attr := range doc.Attributes {
			_, ok := attr.(*tg.DocumentAttributeVideo)
			if ok {
				return true
			}
		}
	}
	return false
}

// Animation returns true if types.Message consists of animation.
func (*messageFilters) Animation(m *types.Message) bool {
	doc := GetDocument(m)
	if doc != nil {
		for _, attr := range doc.Attributes {
			_, ok := attr.(*tg.DocumentAttributeAnimated)
			if ok {
				return true
			}
		}
	}
	return false
}

// Sticker returns true if types.Message consists of sticker.
func (*messageFilters) Sticker(m *types.Message) bool {
	doc := GetDocument(m)
	if doc != nil {
		for _, attr := range doc.Attributes {
			_, ok := attr.(*tg.DocumentAttributeSticker)
			if ok {
				return true
			}
		}
	}
	return false
}

// Audio returns true if types.Message consists of audio.
func (*messageFilters) Audio(m *types.Message) bool {
	doc := GetDocument(m)
	if doc != nil {
		for _, attr := range doc.Attributes {
			_, ok := attr.(*tg.DocumentAttributeAudio)
			if ok {
				return true
			}
		}
	}
	return false
}

// Edited returns true if types.Message is an edited message.
func (*messageFilters) Edited(m *types.Message) bool {
	return m.EditDate != 0
}

func GetDocument(m *types.Message) *tg.Document {
	mdoc, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil
	}
	tgdoc, ok := mdoc.Document.(*tg.Document)
	if !ok {
		return nil
	}
	return tgdoc
}

// Caption returns true if the Message has a caption.
func (*messageFilters) Caption(m *types.Message) bool {
	if mdoc, ok := m.Media.(*tg.MessageMediaDocument); ok {
		return mdoc.Document != nil
	}
	return false
}

// Reply returns true if the Message is a reply to another message.
func (*messageFilters) Reply(m *types.Message) bool {
	return m.ReplyTo != nil
}

// Forwarded returns true if the Message was forwarded from another chat.
func (*messageFilters) Forwarded(m *types.Message) bool {
	return m.FwdFrom.Date != 0
}

// Game returns true if the Message is a game.
func (*messageFilters) Game(m *types.Message) bool {
	game, ok := m.Media.(*tg.MessageMediaGame)
	return ok && game != nil
}

// Poll returns true if the Message contains a poll.
func (*messageFilters) Poll(m *types.Message) bool {
	poll, ok := m.Media.(*tg.MessageMediaPoll)
	return ok && poll != nil
}

// Dice returns true if the Message contains a dice.
func (*messageFilters) Dice(m *types.Message) bool {
	dice, ok := m.Media.(*tg.MessageMediaDice)
	return ok && dice != nil
}

// Voice returns true if the Message contains a voice note.
func (*messageFilters) Voice(m *types.Message) bool {
	doc, ok := m.Media.(*tg.MessageMediaDocument)
	return ok && doc.Voice
}

// VideoNote returns true if the Message contains a video note.
func (*messageFilters) VideoNote(m *types.Message) bool {
	doc, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok {
		return false
	}
	return doc.Round && doc.Video
}

// Contact returns true if the Message contains a contact.
func (*messageFilters) Contact(m *types.Message) bool {
	contact, ok := m.Media.(*tg.MessageMediaContact)
	return ok && contact != nil
}

// Location returns true if the Message contains a location.
func (*messageFilters) Location(m *types.Message) bool {
	location, ok := m.Media.(*tg.MessageMediaGeo)
	return ok && location != nil
}

// Venue returns true if the Message contains a venue.
func (*messageFilters) Venue(m *types.Message) bool {
	venue, ok := m.Media.(*tg.MessageMediaVenue)
	return ok && venue != nil
}

func (*messageFilters) LiveLocation(m *types.Message) bool {
	location, ok := m.Media.(*tg.MessageMediaGeoLive)
	return ok && location != nil
}

func (*messageFilters) Quote(m *types.Message) bool {
	return m.QuickReplyShortcutID != 0
}

// WebPage returns true if the Message has a webpage preview.
func (*messageFilters) WebPage(m *types.Message) bool {
	_, ok := m.Media.(*tg.MessageMediaWebPage)
	return ok
}

// MediaGroup returns true if the Message is part of a media group.
func (*messageFilters) MediaGroup(m *types.Message) bool {
	return m.GroupedID != 0
}

func (*messageFilters) MinLength(minLength int) MessageFilter {
	return func(m *types.Message) bool {
		return len(m.Text) >= minLength
	}
}

func (*messageFilters) MaxLength(maxLength int) MessageFilter {
	return func(m *types.Message) bool {
		return len(m.Text) <= maxLength
	}
}

// Scheduled returns true if the Message was scheduled (sent automatically).
func (*messageFilters) Scheduled(m *types.Message) bool {
	return m.FromScheduled
}

// FromScheduled returns true if the Message was previously scheduled.
func (*messageFilters) FromScheduled(m *types.Message) bool {
	return m.FromScheduled
}

// LinkedChannel returns true if the Message was automatically forwarded from a linked channel.
func (*messageFilters) LinkedChannel(m *types.Message) bool {
	return m.Post && !m.IsOutgoing()
}

// ReplyKeyboard returns true if the Message has a reply keyboard markup.
func (*messageFilters) ReplyKeyboard(m *types.Message) bool {
	_, ok := m.ReplyMarkup.(*tg.ReplyKeyboardMarkup)
	return ok
}

// InlineKeyboard returns true if the Message has an inline keyboard markup.
func (*messageFilters) InlineKeyboard(m *types.Message) bool {
	_, ok := m.ReplyMarkup.(*tg.ReplyInlineMarkup)
	return ok
}

// ViaBot returns true if the Message was sent via inline bot.
func (*messageFilters) ViaBot(m *types.Message) bool {
	return m.ViaBotID != 0
}

// HasMediaSpoiler returns true if the message media has a spoiler.
func (*messageFilters) HasMediaSpoiler(m *types.Message) bool {
	if m.Media == nil {
		return false
	}

	switch media := m.Media.(type) {
	case *tg.MessageMediaDocument:
		return media.Spoiler
	}

	return false
}

// RegexAdvanced returns a Regex filter with options.
func (*messageFilters) RegexAdvanced(pattern string, opts *RegexOptions) MessageFilter {
	capacity := len(pattern) + 12
	var sb strings.Builder
	sb.Grow(capacity)

	if opts != nil {
		if opts.IgnoreCase {
			sb.WriteString("(?i)")
		}
		if opts.Multiline {
			sb.WriteString("(?m)")
		}
		if opts.DotAll {
			sb.WriteString("(?s)")
		}
	}
	sb.WriteString(pattern)

	r := regexp.MustCompile(sb.String())
	return func(m *types.Message) bool {
		return r.MatchString(m.Text)
	}
}

// RegexOptions holds optional parameters for Regex filter.
type RegexOptions struct {
	IgnoreCase bool
	Multiline  bool
	DotAll     bool
}

// MinLength creates a filter that checks if the message text is at least minLength characters long.
func MinLength(minLength int) MessageFilter {
	return func(m *types.Message) bool {
		return (*messageFilters)(nil).MinLength(minLength)(m)
	}
}

// MaxLength creates a filter that checks if the message text is at most maxLength characters long.
func MaxLength(maxLength int) MessageFilter {
	return func(m *types.Message) bool {
		return (*messageFilters)(nil).MaxLength(maxLength)(m)
	}
}

// Document returns true if types.Message consists of a document (generic file).
func (*messageFilters) Document(m *types.Message) bool {
	doc := GetDocument(m)
	if doc == nil {
		return false
	}
	for _, attr := range doc.Attributes {
		switch attr.(type) {
		case *tg.DocumentAttributeVideo,
			*tg.DocumentAttributeAudio,
			*tg.DocumentAttributeAnimated,
			*tg.DocumentAttributeSticker:
			return false
		}
	}
	return true
}

// DocumentFilename returns a filter that matches the document filename against a regex pattern.
func (*messageFilters) DocumentFilename(pattern string) MessageFilter {
	r := regexp.MustCompile(pattern)
	return func(m *types.Message) bool {
		doc := GetDocument(m)
		if doc == nil {
			return false
		}
		for _, attr := range doc.Attributes {
			if fn, ok := attr.(*tg.DocumentAttributeFilename); ok {
				return r.MatchString(fn.FileName)
			}
		}
		return false
	}
}

func (*messageFilters) Me(m *types.Message) bool {
	if m.FromID == nil {
		return false
	}
	peerUser, ok := m.FromID.(*tg.PeerUser)
	return ok && peerUser.UserID == m.PeerID.(*tg.PeerUser).UserID
}

func (*messageFilters) Bot(m *types.Message) bool {
	if m.FromID == nil {
		return false
	}
	peerUser, ok := m.FromID.(*tg.PeerUser)
	return ok && peerUser.UserID == 0
}

func (*messageFilters) SenderChat(m *types.Message) bool {
	return m.SavedPeerID != nil
}

func (*messageFilters) Incoming(m *types.Message) bool {
	return !m.IsOutgoing()
}

func (*messageFilters) Outgoing(m *types.Message) bool {
	return m.IsOutgoing()
}

func (*messageFilters) SelfDestruction(m *types.Message) bool {
	if m.Media == nil {
		return false
	}
	if doc, ok := m.Media.(*tg.MessageMediaDocument); ok && doc.Document != nil {
		return doc.TTLSeconds != 0
	}
	return false
}

func (*messageFilters) Private(m *types.Message) bool {
	_, ok := m.PeerID.(*tg.PeerUser)
	return ok
}

func (*messageFilters) Group(m *types.Message) bool {
	_, isChat := m.PeerID.(*tg.PeerChat)
	_, isChannel := m.PeerID.(*tg.PeerChannel)
	return isChat || isChannel
}

func (*messageFilters) Channel(m *types.Message) bool {
	_, ok := m.PeerID.(*tg.PeerChannel)
	return ok
}

func (*messageFilters) Story(m *types.Message) bool {
	return false
}
