package types

import (
	"github.com/gotd/td/tg"
)

// Photo returns the photo if message contains photo media, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Photo() *Photo {
	if m == nil || m.Message == nil || m.Media == nil {
		return nil
	}
	if media, ok := m.Media.(*tg.MessageMediaPhoto); ok {
		if photo, ok := media.Photo.AsNotEmpty(); ok {
			return NewPhoto(photo)
		}
	}
	return nil
}

// Document returns the document if message contains document media, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Document() *Document {
	if m == nil || m.Message == nil || m.Media == nil {
		return nil
	}
	if media, ok := m.Media.(*tg.MessageMediaDocument); ok {
		if doc, ok := media.Document.(*tg.Document); ok {
			return NewDocument(doc)
		}
	}
	return nil
}

// Geo returns the location if message contains geo media, nil otherwise.
func (m *Message) Geo() *tg.MessageMediaGeo {
	if m.Media == nil {
		return nil
	}
	if geo, ok := m.Media.(*tg.MessageMediaGeo); ok {
		return geo
	}
	return nil
}

// GeoLive returns the live location if message contains geo live media, nil otherwise.
func (m *Message) GeoLive() *tg.MessageMediaGeoLive {
	if m.Media == nil {
		return nil
	}
	if geoLive, ok := m.Media.(*tg.MessageMediaGeoLive); ok {
		return geoLive
	}
	return nil
}

// Contact returns the contact if message contains contact media, nil otherwise.
func (m *Message) Contact() *tg.MessageMediaContact {
	if m.Media == nil {
		return nil
	}
	if contact, ok := m.Media.(*tg.MessageMediaContact); ok {
		return contact
	}
	return nil
}

// Poll returns the poll if message contains poll media, nil otherwise.
func (m *Message) Poll() *tg.MessageMediaPoll {
	if m.Media == nil {
		return nil
	}
	if poll, ok := m.Media.(*tg.MessageMediaPoll); ok {
		return poll
	}
	return nil
}

// Venue returns the venue if message contains venue media, nil otherwise.
func (m *Message) Venue() *tg.MessageMediaVenue {
	if m.Media == nil {
		return nil
	}
	if venue, ok := m.Media.(*tg.MessageMediaVenue); ok {
		return venue
	}
	return nil
}

// Dice returns the dice if message contains dice media, nil otherwise.
func (m *Message) Dice() *tg.MessageMediaDice {
	if m.Media == nil {
		return nil
	}
	if dice, ok := m.Media.(*tg.MessageMediaDice); ok {
		return dice
	}
	return nil
}

// Game returns the game if message contains game media, nil otherwise.
func (m *Message) Game() *tg.MessageMediaGame {
	if m.Media == nil {
		return nil
	}
	if game, ok := m.Media.(*tg.MessageMediaGame); ok {
		return game
	}
	return nil
}

// WebPage returns the webpage if message contains webpage media, nil otherwise.
func (m *Message) WebPage() *tg.MessageMediaWebPage {
	if m.Media == nil {
		return nil
	}
	if wp, ok := m.Media.(*tg.MessageMediaWebPage); ok {
		return wp
	}
	return nil
}

// Invoice returns the invoice if message contains invoice media, nil otherwise.
func (m *Message) Invoice() *tg.MessageMediaInvoice {
	if m.Media == nil {
		return nil
	}
	if invoice, ok := m.Media.(*tg.MessageMediaInvoice); ok {
		return invoice
	}
	return nil
}

// Giveaway returns the giveaway if message contains giveaway media, nil otherwise.
func (m *Message) Giveaway() *tg.MessageMediaGiveaway {
	if m.Media == nil {
		return nil
	}
	if giveaway, ok := m.Media.(*tg.MessageMediaGiveaway); ok {
		return giveaway
	}
	return nil
}

// GiveawayResults returns the giveaway results if message contains giveaway results media, nil otherwise.
func (m *Message) GiveawayResults() *tg.MessageMediaGiveawayResults {
	if m.Media == nil {
		return nil
	}
	if results, ok := m.Media.(*tg.MessageMediaGiveawayResults); ok {
		return results
	}
	return nil
}

// Story returns the story if message contains story media, nil otherwise.
func (m *Message) Story() *tg.MessageMediaStory {
	if m.Media == nil {
		return nil
	}
	if story, ok := m.Media.(*tg.MessageMediaStory); ok {
		return story
	}
	return nil
}

// PaidMedia returns the paid media if message contains paid media, nil otherwise.
func (m *Message) PaidMedia() *tg.MessageMediaPaidMedia {
	if m.Media == nil {
		return nil
	}
	if paid, ok := m.Media.(*tg.MessageMediaPaidMedia); ok {
		return paid
	}
	return nil
}

// Video returns the document if message contains a video, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Video() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
			return doc
		}
	}
	return nil
}

// Audio returns the document if message contains an audio file, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Audio() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeAudio); ok {
			return doc
		}
	}
	return nil
}

// Voice returns the document if message contains a voice note, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Voice() *Document {
	if m.Media == nil {
		return nil
	}
	media, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok || !media.Voice {
		return nil
	}
	if doc, ok := media.Document.(*tg.Document); ok {
		return NewDocument(doc)
	}
	return nil
}

// Animation returns the document if message contains an animation (GIF), nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Animation() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeAnimated); ok {
			return doc
		}
	}
	return nil
}

// VideoNote returns the document if message contains a video note (round video), nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) VideoNote() *Document {
	if m.Media == nil {
		return nil
	}
	media, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil
	}
	if !media.Round || !media.Video {
		return nil
	}
	if doc, ok := media.Document.(*tg.Document); ok {
		return NewDocument(doc)
	}
	return nil
}

// Sticker returns the document if message contains a sticker, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
func (m *Message) Sticker() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeSticker); ok {
			return doc
		}
	}
	return nil
}

// IsMedia returns true if the message contains any media.
func (m *Message) IsMedia() bool {
	return m.Media != nil
}

// FileID returns a formatted file ID string for any media in the message.
// This is a convenience method that calls FileID() on the appropriate media wrapper.
// Returns empty string if the message contains no media.
func (m *Message) FileID() string {
	if photo := m.Photo(); photo != nil {
		return photo.FileID()
	}

	if doc := m.Document(); doc != nil {
		return doc.FileID()
	}

	return ""
}
