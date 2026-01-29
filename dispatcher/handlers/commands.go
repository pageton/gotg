package handlers

import (
	"strings"
	"unicode"

	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// Command handler is executed when the update consists of tg.Message provided it is a command and satisfies all the conditions.
type Command struct {
	Prefix        []rune
	Name          string
	Callback      CallbackResponse
	Outgoing      bool
	UpdateFilters filters.UpdateFilter
}

// DefaultPrefix is the global variable consisting all the prefixes which will trigger the command.
var DefaultPrefix = []rune{'!', '/'}

// NewCommand creates a new Command handler with default fields, bound to call its response
func NewCommand(name string, response CallbackResponse) Command {
	return Command{
		Name:     name,
		Callback: response,
		Prefix:   DefaultPrefix,
		Outgoing: true,
	}
}

// OnCommand creates a new Command handler with an UpdateHandler.
// This is a convenience function for handlers that only need the Update parameter.
func OnCommand(name string, handler UpdateHandler, updateFilters ...filters.UpdateFilter) Command {
	updateFilter := func(u *adapter.Update) bool {
		if len(updateFilters) == 0 {
			return true
		}
		for _, f := range updateFilters {
			if !f(u) {
				return false
			}
		}
		return true
	}
	return Command{
		Name:          name,
		Callback:      ToCallbackResponse(handler),
		Prefix:        DefaultPrefix,
		Outgoing:      true,
		UpdateFilters: updateFilter,
	}
}

func (c Command) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	m := u.EffectiveMessage
	if m == nil || m.Text == "" {
		return nil
	}
	if !c.Outgoing && m.IsOutgoing() {
		return nil
	}
	if c.UpdateFilters != nil && !c.UpdateFilters(u) {
		return nil
	}

	// Optimized: Parse command without allocating new strings
	// Instead of strings.Fields() + strings.ToLower(), do inline parsing
	text := m.Text
	if len(text) == 0 {
		return nil
	}

	// Find first space (end of first word/command)
	end := len(text)
	for i, r := range text {
		if unicode.IsSpace(r) {
			end = i
			break
		}
	}

	// Extract first word and convert to lowercase in-place
	firstWord := text[:end]
	arg := toLowerInline(firstWord)

	// Check each prefix
	for _, prefix := range c.Prefix {
		if len(arg) > 0 && arg[0] == byte(prefix) {
			cmdPart := arg[1:]
			if cmdPart == c.Name {
				return c.Callback(ctx, u)
			}
			// Check for username-qualified command (e.g., /start@botname)
			if atIdx := strings.IndexByte(cmdPart, '@'); atIdx > 0 {
				if cmdPart[:atIdx] == c.Name {
					username := cmdPart[atIdx+1:]
					if username == strings.ToLower(ctx.Self.Username) {
						return c.Callback(ctx, u)
					}
				}
			}
		}
	}
	return nil
}

// toLowerInline converts ASCII string to lowercase without allocations
// For strings that may contain non-ASCII, it only lowercases ASCII characters
func toLowerInline(s string) string {
	// Fast path: check if string needs conversion
	needsConversion := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			needsConversion = true
			break
		}
	}

	if !needsConversion {
		return s
	}

	// Convert in-place using a byte slice
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32 // 'a' - 'A' = 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
