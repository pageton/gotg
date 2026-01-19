// Package parsemode provides robust parsing for Telegram message formatting.
// It supports HTML, MarkdownV2, and None parse modes with thread-safe operations.
//
// Basic usage:
//
//	result, err := Parse("<b>Hello</b>", ModeHTML)
//	if err != nil {
//	    // handle error
//	}
//	text := result.Text
//	entities := result.Entities
//
// The package is thread-safe and all parsers can be used concurrently.
package parsemode

import (
	"sync"

	"github.com/gotd/td/tg"
)

// Default parsers instance (singleton pattern for efficiency)
var (
	defaultHTMLParser     *HTMLParser
	defaultMarkdownParser *MarkdownParser
	initOnce             sync.Once
)

func initParsers() {
	if defaultHTMLParser == nil {
		defaultHTMLParser = NewHTMLParser()
	}
	if defaultMarkdownParser == nil {
		defaultMarkdownParser = NewMarkdownParser()
	}
}

// Parse parses the input text according to the given parse mode.
// This is the main entry point for the package.
//
// Parameters:
//   - input: The formatted text to parse
//   - mode: The parse mode (ModeHTML, ModeMarkdown, or ModeNone)
//
// Returns:
//   - *ParseResult: The parsed text and entities
//   - error: Any parsing error (currently always returns nil on parse errors,
//     instead returning the original text with no entities)
func Parse(input string, mode ParseMode) (*ParseResult, error) {
	initOnce.Do(initParsers)

	switch mode {
	case ModeHTML:
		return defaultHTMLParser.Parse(input)
	case ModeMarkdown:
		return defaultMarkdownParser.Parse(input)
	case ModeNone:
		return &ParseResult{Text: input, Entities: nil}, nil
	default:
		// Unknown mode, treat as ModeNone
		return &ParseResult{Text: input, Entities: nil}, nil
	}
}

// ParseHTML parses HTML formatted text.
func ParseHTML(input string) (*ParseResult, error) {
	initOnce.Do(initParsers)
	return defaultHTMLParser.Parse(input)
}

// ParseMarkdown parses MarkdownV2 formatted text.
func ParseMarkdown(input string) (*ParseResult, error) {
	initOnce.Do(initParsers)
	return defaultMarkdownParser.Parse(input)
}

// FormatText formats the input text according to the given parse mode.
// For ModeNone, this returns the text as-is.
// For ModeHTML, this returns the text as-is (no formatting needed).
// For ModeMarkdown, this escapes special characters.
func FormatText(input string, mode ParseMode) string {
	switch mode {
	case ModeHTML, ModeNone:
		return input
	case ModeMarkdown:
		initOnce.Do(initParsers)
		return defaultMarkdownParser.Format(input)
	default:
		return input
	}
}

// TextToEntities converts text with entities to the format expected by Telegram API.
// This is a convenience function for sending messages.
func TextToEntities(text string, entities []tg.MessageEntityClass) (string, []tg.MessageEntityClass) {
	// Currently just returns the input as-is
	// In the future, this could validate and normalize entities
	return text, entities
}

// EntitiesToText converts Telegram entities back to formatted text.
// This is useful for displaying received messages with formatting.
func EntitiesToText(text string, entities []tg.MessageEntityClass, mode ParseMode) string {
	if mode == ModeNone || len(entities) == 0 {
		return text
	}

	if mode == ModeHTML {
		// Would need to implement entity to HTML conversion
		// For now, return plain text
		return text
	}

	// For other modes, we'd need to implement conversion
	// For now, return plain text
	return text
}

// StripFormatting removes all formatting from text.
// This is useful for getting plain text from formatted messages.
func StripFormatting(text string, entities []tg.MessageEntityClass) string {
	// Simply return the text as-is
	// The entities are metadata, not part of the actual text
	return text
}

// IsValidParseMode checks if the given parse mode is valid.
func IsValidParseMode(mode ParseMode) bool {
	return mode.IsValid()
}

// GetParseModeFromString converts a string to a ParseMode.
// Returns ModeNone if the string is not recognized.
func GetParseModeFromString(s string) ParseMode {
	switch s {
	case "", "none":
		return ModeNone
	case "HTML", "html":
		return ModeHTML
	case "MarkdownV2", "markdown", "Markdown":
		return ModeMarkdown
	default:
		return ModeNone
	}
}

// ParseModeFromEntities determines the best parse mode from existing entities.
// This is useful when you want to re-format a message.
func ParseModeFromEntities(entities []tg.MessageEntityClass) ParseMode {
	// If no entities, use ModeNone
	if len(entities) == 0 {
		return ModeNone
	}

	// For simplicity, default to ModeHTML
	// A more sophisticated implementation could analyze the entities
	return ModeHTML
}
