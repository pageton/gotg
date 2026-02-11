package gotg

import (
	"github.com/gotd/td/tg"
)

// Next moves to the next row in the keyboard layout.
//
// Telegram keyboards are organized into rows of buttons. The Next() method
// finalizes the current row and starts a new empty row. Any buttons added
// after calling Next() will appear in the next row of the keyboard.
//
// If the current row is empty (no buttons added since the last Next() or
// since the builder was created), this method does nothing and returns the
// builder unchanged.
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Button("Row 1, Col 1", "r1c1").
//	    Button("Row 1, Col 2", "r1c2").
//	    Next().  // Move to row 2
//	    Button("Row 2, Col 1", "r2c1").
//	    Next().  // Move to row 3
//	    Button("Row 3, Col 1", "r3c1").
//	    Button("Row 3, Col 2", "r3c2").
//	    Button("Row 3, Col 3", "r3c3").
//	    Build()
//
// Note: The Build() and BuildReply() methods automatically call Next() to
// finalize the last row, so you don't need to call it explicitly before building.
func (k *KeyboardBuilder) Next() *KeyboardBuilder {
	if len(k.row) > 0 {
		k.rows = append(k.rows, k.row)
		k.row = make([]tg.KeyboardButtonClass, 0)
	}
	return k
}

// buildRows finalizes the current row and returns all rows for building.
//
// This is an internal method that:
// 1. Calls Next() to finalize the current row
// 2. Returns nil if no rows were added (empty keyboard)
// 3. Returns the accumulated rows otherwise
//
// Returns a 2D slice of keyboard button rows, or nil if no buttons were added.
func (k *KeyboardBuilder) buildRows() [][]tg.KeyboardButtonClass {
	k.Next()
	if len(k.rows) == 0 {
		return nil
	}
	return k.rows
}

// Build constructs an inline keyboard from the configured buttons.
//
// Inline keyboards appear directly below the message and remain visible as the
// user scrolls. They are ideal for contextual actions, navigation, and interactive
// interfaces. Inline keyboards support callback buttons, URL buttons, and other
// inline-specific button types.
//
// The returned tg.ReplyMarkupClass can be used when sending or editing messages
// via the reply markup parameter.
//
// Returns a tg.ReplyInlineMarkup containing the configured rows, or nil if no
// buttons were added to the builder.
//
// Example:
//
//	// Build an inline keyboard with callback and URL buttons
//	keyboard := gotg.Keyboard().
//	    Button("View Details", "view_123").
//	    Button("Delete", "delete_123").
//	    Next().
//	    URL("Website", "https://example.com").
//	    Next().
//	    Switch("Share", false, "share_123").
//	    Build()
//
//	// Use the keyboard when sending a message
//	_, err := ctx.SendMessage(ctx, &tg.MessagesSendMessageRequest{
//	    Peer:        peer,
//	    Message:     "Choose an action:",
//	    ReplyMarkup: keyboard,
//	})
//
// Supported button types for inline keyboards:
// - Button() - Callback buttons (trigger OnCallbackQuery handlers)
// - URL() - Open a web link
// - Switch() - Switch to inline mode
// - Copy() - Copy text to clipboard
// - Game() - Launch an HTML5 game
// - Buy() - Payment button
// - WebApp() - Open a Telegram Mini App
// - Text() - Send text as a message (not recommended, use Button instead)
//
// Note: For reply keyboards (which appear above the input field), use BuildReply()
// instead. RequestPhone, RequestGeo, and RequestPoll buttons are not supported in
// inline keyboards.
func (k *KeyboardBuilder) Build() tg.ReplyMarkupClass {
	rows := k.buildRows()
	if rows == nil {
		return nil
	}
	return &tg.ReplyInlineMarkup{
		Rows: makeRows(rows),
	}
}

// ReplyOptions defines optional configuration for reply keyboards.
//
// Reply keyboards appear above the message input field in the Telegram app
// and provide quick access to common commands or actions. These options control
// the keyboard's appearance and behavior.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Text("Option 1").
//	    Text("Option 2").
//	    BuildReply(gotg.ReplyOptions{
//	        Resize:      true,
//	        OneTime:     true,
//	        Placeholder: "Select an option...",
//	    })
type ReplyOptions struct {
	// Resize controls whether the keyboard is resized vertically to fit its buttons.
	// When true, the keyboard will be smaller and take up less screen space.
	// This is recommended for keyboards with few buttons.
	// Default: false (keyboard uses full height)
	Resize bool

	// OneTime controls whether the keyboard hides after one button press.
	// When true, the keyboard disappears immediately after the user taps a button.
	// This is useful for one-time selections like choosing a language or confirming an action.
	// Default: false (keyboard remains visible)
	OneTime bool

	// Selective controls whether the keyboard is shown only to specific users.
	// When true, the keyboard is shown only to users mentioned in the message's
	// replied-to message or whose username is mentioned in the text.
	// Default: false (keyboard shown to all users)
	Selective bool

	// Persistent controls whether the keyboard remains visible even when a message
	// with a different keyboard is sent.
	// When true, this keyboard will not be automatically hidden by other keyboards.
	// Default: false (keyboard is replaced by new keyboards)
	Persistent bool

	// Placeholder sets the placeholder text shown in the input field when the
	// keyboard is active. This provides a hint or prompt to the user.
	// Empty string (default) uses the standard placeholder.
	Placeholder string
}

// BuildReply constructs a reply keyboard from the configured buttons.
//
// Reply keyboards appear above the message input field in the Telegram app and
// provide quick access to commands or actions. Unlike inline keyboards, reply
// keyboard buttons send their text as regular messages when tapped.
//
// The returned tg.ReplyMarkupClass can be used when sending or editing messages
// via the reply markup parameter. To hide a reply keyboard, send a keyboard with
// ReplyKeyboardForceEmpty or use context.RemoveReplyKeyboard().
//
// Parameters:
//   - opts: Optional ReplyOptions to configure keyboard behavior (resize, one-time,
//     etc.). If omitted, default options are used.
//
// Returns a tg.ReplyKeyboardMarkup containing the configured rows and options,
// or nil if no buttons were added to the builder.
//
// Example:
//
//	// Basic reply keyboard
//	keyboard := gotg.Keyboard().
//	    Text("Start").
//	    Text("Help").
//	    Next().
//	    Text("Settings").
//	    BuildReply()
//
//	// Reply keyboard with options
//	keyboard := gotg.Keyboard().
//	    Text("Yes").
//	    Text("No").
//	    RequestPhone("Share Phone").
//	    BuildReply(gotg.ReplyOptions{
//	        Resize:      true,    // Smaller keyboard
//	        OneTime:     true,    // Hide after selection
//	        Placeholder: "Choose: Yes or No",
//	    })
//
//	// Use the keyboard when sending a message
//	_, err := ctx.SendMessage(ctx, &tg.MessagesSendMessageRequest{
//	    Peer:        peer,
//	    Message:     "Please choose an option:",
//	    ReplyMarkup: keyboard,
//	})
//
// Supported button types for reply keyboards:
// - Text() - Send text as a message
// - RequestPhone() - Request the user's phone number
// - RequestGeo() - Request the user's location
// - RequestPoll() - Create a poll or quiz
// - WebApp() - Open a Telegram Mini App
//
// Note: For inline keyboards (which appear below the message), use Build() instead.
// Callback buttons (Button()), URL buttons (URL()), and Switch buttons are not
// supported in reply keyboards.
//
// To hide a reply keyboard after it's been shown:
//
//	_, err := ctx.SendMessage(ctx, &tg.MessagesSendMessageRequest{
//	    Peer:        peer,
//	    Message:     "Keyboard hidden",
//	    ReplyMarkup: &tg.ReplyKeyboardHide{},
//	})
func (k *KeyboardBuilder) BuildReply(opts ...ReplyOptions) tg.ReplyMarkupClass {
	rows := k.buildRows()
	if rows == nil {
		return nil
	}

	options := ReplyOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	markup := &tg.ReplyKeyboardMarkup{
		Rows:       makeRows(rows),
		Resize:     options.Resize,
		SingleUse:  options.OneTime,
		Selective:  options.Selective,
		Persistent: options.Persistent,
	}

	if options.Placeholder != "" {
		markup.Placeholder = options.Placeholder
	}

	return markup
}

// makeRows converts a 2D slice of keyboard button classes to the format
// required by Telegram's API.
//
// This is an internal helper function that transforms each row slice into a
// tg.KeyboardButtonRow struct, which wraps the buttons for API transmission.
//
// Parameters:
//   - rows: A 2D slice where each inner slice represents a row of buttons
//
// Returns a slice of tg.KeyboardButtonRow suitable for use in ReplyInlineMarkup
// or ReplyKeyboardMarkup.
func makeRows(rows [][]tg.KeyboardButtonClass) []tg.KeyboardButtonRow {
	result := make([]tg.KeyboardButtonRow, len(rows))
	for i, row := range rows {
		result[i] = tg.KeyboardButtonRow{Buttons: row}
	}
	return result
}
