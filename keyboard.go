package gotg

import (
	"github.com/gotd/td/tg"
)

// KeyboardBuilder provides a fluent interface for constructing Telegram keyboards.
//
// KeyboardBuilder supports both inline keyboards (for callback buttons, URLs, etc.)
// and reply keyboards (for regular text buttons that appear above the input field).
//
// The builder uses a row-based approach where buttons are added to the current row,
// and the Next() method moves to the next row. The final keyboard is built using
// either Build() for inline keyboards or BuildReply() for reply keyboards.
//
// Example (inline keyboard):
//
//	keyboard := gotg.Keyboard().
//	    Button("Button 1", "data1").
//	    Button("Button 2", "data2").
//	    Next().
//	    URL("Visit", "https://example.com").
//	    Build()
//
// Example (reply keyboard):
//
//	keyboard := gotg.Keyboard().
//	    Text("Option 1").
//	    Text("Option 2").
//	    Next().
//	    RequestPhone("Share Phone").
//	    BuildReply(gotg.ReplyOptions{Resize: true})
type KeyboardBuilder struct {
	rows [][]tg.KeyboardButtonClass
	row  []tg.KeyboardButtonClass
}

// Keyboard creates a new KeyboardBuilder instance for constructing Telegram keyboards.
//
// The returned builder is initialized with empty rows and can be used to build
// both inline and reply keyboards through method chaining.
//
// Returns a new KeyboardBuilder ready for button construction.
//
// Example:
//
//	builder := gotg.Keyboard()
//	keyboard := builder.
//	    Button("Click me", "callback_data").
//	    Next().
//	    Text("Another").
//	    Build()
func Keyboard() *KeyboardBuilder {
	return &KeyboardBuilder{
		rows: make([][]tg.KeyboardButtonClass, 0),
		row:  make([]tg.KeyboardButtonClass, 0),
	}
}

// add adds a button to the current row of the keyboard.
//
// This is an internal method that appends the given button to the current row
// and returns the builder for method chaining. The button will be part of the
// current row until Next() is called to start a new row.
func (k *KeyboardBuilder) add(btn tg.KeyboardButtonClass) *KeyboardBuilder {
	k.row = append(k.row, btn)
	return k
}

// Text adds a simple text button to the keyboard.
//
// Text buttons send their text as a message when tapped. These buttons are
// primarily used in reply keyboards (built with BuildReply), as inline keyboards
// should use Button() for callback buttons instead.
//
// When used in an inline keyboard (built with Build), text buttons will still
// function but will send the text as a regular message from the user.
//
// Parameters:
//   - text: The label text to display on the button (and the message text when clicked)
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Text("Yes").
//	    Text("No").
//	    Next().
//	    Text("Maybe").
//	    BuildReply()
func (k *KeyboardBuilder) Text(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButton{Text: text})
}

// Button adds a callback button with data payload to the keyboard.
//
// Callback buttons are designed for inline keyboards (built with Build). When
// tapped, they trigger a callback query that your bot can handle via OnCallbackQuery.
// The data parameter is passed back to your bot, allowing you to identify which
// button was clicked and take appropriate action.
//
// The data parameter is limited to 64 bytes in Telegram's API.
//
// Parameters:
//   - text: The label text to display on the button
//   - data: The callback data to send when the button is tapped (max 64 bytes)
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Button("Approve", "approve_123").
//	    Button("Reject", "reject_123").
//	    Next().
//	    Button("View Details", "details_123").
//	    Build()
//
// Handling the callback:
//
//	dispatcher.OnCallbackQuery(func(ctx *context.Context, u *context.Update) error {
//	    data := string(u.CallbackQuery.Data)
//	    // data will be "approve_123", "reject_123", or "details_123"
//	    return u.Answer("Received: " + data)
//	})
func (k *KeyboardBuilder) Button(text, data string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonCallback{
		Text: text,
		Data: []byte(data),
	})
}

// URL adds a button that opens a URL when tapped.
//
// URL buttons are designed for inline keyboards. When the user taps the button,
// Telegram will open the specified URL in the user's browser. The button text
// is displayed as a clickable link.
//
// These buttons are ideal for directing users to external resources such as
// websites, documentation, or payment pages.
//
// Parameters:
//   - text: The label text to display on the button
//   - url: The URL to open when the button is tapped (must be valid)
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    URL("Visit Website", "https://example.com").
//	    URL("Documentation", "https://docs.example.com").
//	    Next().
//	    URL("Support", "https://t.me/supportbot").
//	    Build()
//
// Note: HTTPS URLs are preferred. Telegram may show warnings for HTTP URLs.
func (k *KeyboardBuilder) URL(text, url string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonURL{
		Text: text,
		URL:  url,
	})
}

// Switch adds a button that switches to inline mode when tapped.
//
// Switch buttons are designed for inline keyboards. When tapped, they prompt
// the user to select a chat where the inline query will be sent. Your bot will
// receive an inline query via OnInlineQuery that can be used to provide
// contextual results.
//
// This is commonly used for features like "Share with friends" or
// "Send to another chat".
//
// Parameters:
//   - text: The label text to display on the button
//   - samePeer: If true, the inline query will be sent in the current chat;
//     if false, the user can select any chat
//   - query: The query string to pre-fill in the inline input
//
// Returns the builder for method chaining.
//
// Example:
//
//	// Button allows sharing content in the current chat
//	keyboard := gotg.Keyboard().
//	    Switch("Share here", true, "share_item_123").
//	    Next().
//	    // Button allows selecting any chat to share to
//	    Switch("Share with friends", false, "share_item_123").
//	    Build()
//
// Handling the inline query:
//
//	dispatcher.OnInlineQuery(func(ctx *context.Context, u *context.Update) error {
//	    query := u.InlineQuery.Query()
//	    // query will contain "share_item_123" plus any user modifications
//	    return u.AnswerInlineQuery(results, nil)
//	})
func (k *KeyboardBuilder) Switch(text string, samePeer bool, query string) *KeyboardBuilder {
	btn := &tg.KeyboardButtonSwitchInline{
		Text:     text,
		Query:    query,
		SamePeer: samePeer,
	}
	return k.add(btn)
}
