// Package parsemode provides robust parsing for Telegram message formatting.
// It supports HTML, MarkdownV2, and None parse modes with thread-safe operations.
package parsemode

import (
	"github.com/gotd/td/tg"
)

// ParseMode represents the type of formatting to apply to a message.
type ParseMode string

const (
	// ModeNone indicates no formatting should be applied.
	ModeNone ParseMode = ""
	// ModeHTML indicates HTML formatting should be parsed.
	ModeHTML ParseMode = "HTML"
	// ModeMarkdown indicates MarkdownV2 formatting should be parsed.
	ModeMarkdown ParseMode = "MarkdownV2"
)

// String returns the string representation of the parse mode.
func (p ParseMode) String() string {
	return string(p)
}

// IsValid returns true if the parse mode is valid.
func (p ParseMode) IsValid() bool {
	switch p {
	case ModeNone, ModeHTML, ModeMarkdown:
		return true
	}
	return false
}

// EntityType represents the type of a message entity.
type EntityType string

const (
	EntityTypeBold               EntityType = "bold"
	EntityTypeItalic             EntityType = "italic"
	EntityTypeUnderline          EntityType = "underline"
	EntityTypeStrike             EntityType = "strike"
	EntityTypeSpoiler            EntityType = "spoiler"
	EntityTypeCode               EntityType = "code"
	EntityTypePre                EntityType = "pre"
	EntityTypeTextURL            EntityType = "text_url"
	EntityTypeURL                EntityType = "url"
	EntityTypeEmail              EntityType = "email"
	EntityTypeMention            EntityType = "mention"
	EntityTypeMentionName        EntityType = "mention_name"
	EntityTypeCustomEmoji        EntityType = "custom_emoji"
	EntityTypeBlockquote         EntityType = "blockquote"
	EntityTypeTextAnchor         EntityType = "text_anchor"
	EntityTypeCashtag            EntityType = "cashtag"
	EntityTypeHashtag            EntityType = "hashtag"
	EntityTypeBotCommand         EntityType = "bot_command"
	EntityTypePhone              EntityType = "phone"
	EntityTypeBankCard           EntityType = "bank_card"
	EntityTypeExpandedBlockquote EntityType = "expanded_blockquote"
)

// ParseResult contains the parsed text and entities.
type ParseResult struct {
	// Text is the cleaned/parsed text content.
	Text string
	// Entities are the parsed message entities.
	Entities []tg.MessageEntityClass
}

// Parser is the interface for parsing formatted text into Telegram entities.
type Parser interface {
	// Parse parses the input text and returns the result with entities.
	Parse(input string) (*ParseResult, error)
}

// TextFormatter formats text with the given parse mode.
type TextFormatter interface {
	// Format formats the input text according to the parse mode.
	Format(input string) string
}

// EntityBuilder builds Telegram message entities.
type EntityBuilder struct {
	entities []tg.MessageEntityClass
}

// NewEntityBuilder creates a new entity builder.
func NewEntityBuilder() *EntityBuilder {
	return &EntityBuilder{
		entities: make([]tg.MessageEntityClass, 0),
	}
}

// Add adds an entity to the builder.
func (b *EntityBuilder) Add(entity tg.MessageEntityClass) {
	b.entities = append(b.entities, entity)
}

// Build returns the built entities slice.
func (b *EntityBuilder) Build() []tg.MessageEntityClass {
	if len(b.entities) == 0 {
		return nil
	}
	return b.entities
}

// Reset clears all entities from the builder.
func (b *EntityBuilder) Reset() {
	b.entities = b.entities[:0]
}

// Clone creates a deep copy of the entity builder.
func (b *EntityBuilder) Clone() *EntityBuilder {
	cloned := &EntityBuilder{
		entities: make([]tg.MessageEntityClass, len(b.entities)),
	}
	copy(cloned.entities, b.entities)
	return cloned
}
