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
//	    query := u.InlineQuery.Query
//	    // query will contain "share_item_123" plus any user modifications
//	    return u.Answer(&types.InlineQueryResultArticle{...})
//	})
func (k *KeyboardBuilder) Switch(text string, samePeer bool, query string) *KeyboardBuilder {
	btn := &tg.KeyboardButtonSwitchInline{
		Text:     text,
		Query:    query,
		SamePeer: samePeer,
	}
	return k.add(btn)
}

// Copy adds a button that copies text to the user's clipboard when tapped.
//
// Copy buttons are designed for inline keyboards. When tapped, Telegram copies
// the specified text to the user's clipboard and shows a confirmation notification.
// The button text remains on the screen and can be tapped multiple times.
//
// This is useful for sharing codes, tokens, addresses, or any text that the user
// might need to paste elsewhere.
//
// Parameters:
//   - text: The label text to display on the button
//   - copyText: The text that will be copied to the clipboard when tapped
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Copy("Copy invite link", "https://t.me/examplebot?start=abc123").
//	    Copy("Copy promo code", "SAVE20").
//	    Next().
//	    Copy("Copy address", "0x1234567890abcdef").
//	    Build()
//
// Note: The copyText can be different from the button label, allowing you to
// display user-friendly text while copying technical content like URLs or codes.
func (k *KeyboardBuilder) Copy(text, copyText string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonCopy{
		Text:     text,
		CopyText: copyText,
	})
}

// RequestPhone adds a button that requests the user's phone number.
//
// Phone request buttons are designed for reply keyboards (built with BuildReply).
// When tapped, the user will be prompted to share their phone number with the bot.
// The user can choose to share their phone number or cancel the request.
//
// These buttons are commonly used for collecting user contact information during
// registration or verification flows. Telegram handles the phone number input UI,
// ensuring a consistent user experience.
//
// Only one phone request button can exist per keyboard. Adding more than one may
// result in undefined behavior.
//
// Parameters:
//   - text: The label text to display on the button (e.g., "Share Phone Number")
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Text("Skip").
//	    RequestPhone("Share Phone Number").
//	    BuildReply(gotg.ReplyOptions{Resize: true})
//
// Handling the phone number:
//
//	dispatcher.OnMessage(func(ctx *context.Context, u *context.Update) error {
//	    if m, ok := u.EffectiveMessage(); ok {
//	        if contact := m.Contact(); ok {
//	            phone := contact.PhoneNumber
//	            // Process the phone number
//	        }
//	    }
//	    return nil
//	})
func (k *KeyboardBuilder) RequestPhone(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonRequestPhone{Text: text})
}

// RequestGeo adds a button that requests the user's geolocation.
//
// Location request buttons are designed for reply keyboards (built with BuildReply).
// When tapped, the user will be prompted to share their current location with the bot.
// The user can choose to share their location or cancel the request.
//
// These buttons are commonly used for location-based features such as finding nearby
// places, delivery services, or location-aware content. Telegram handles the location
// request UI, ensuring a consistent user experience.
//
// Only one location request button can exist per keyboard. Adding more than one may
// result in undefined behavior.
//
// Parameters:
//   - text: The label text to display on the button (e.g., "Send Location")
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Text("Enter manually").
//	    RequestGeo("Send Location").
//	    BuildReply(gotg.ReplyOptions{Resize: true})
//
// Handling the location:
//
//	dispatcher.OnMessage(func(ctx *context.Context, u *context.Update) error {
//	    if m, ok := u.EffectiveMessage(); ok {
//	        if geo := m.Geo(); ok {
//	            lat := geo.Lat
//	            long := geo.Long
//	            // Process the coordinates
//	        }
//	    }
//	    return nil
//	})
func (k *KeyboardBuilder) RequestGeo(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonRequestGeoLocation{Text: text})
}

// RequestPoll adds a button that creates a poll when tapped.
//
// Poll request buttons are designed for reply keyboards (built with BuildReply).
// When tapped, the user will be prompted to create a poll or quiz and send it
// to the chat. Telegram provides a full UI for creating the poll with questions,
// options, and settings.
//
// These buttons are useful for engaging users and collecting feedback or opinions
// in a structured format.
//
// Only one poll request button can exist per keyboard. Adding more than one may
// result in undefined behavior.
//
// Parameters:
//   - text: The label text to display on the button
//   - quiz: If true, creates a quiz (poll with correct answers);
//     if false, creates a regular poll
//
// Returns the builder for method chaining.
//
// Example:
//
//	// Regular poll button
//	keyboard := gotg.Keyboard().
//	    RequestPoll("Create Poll", false).
//	    BuildReply(gotg.ReplyOptions{Resize: true})
//
//	// Quiz button
//	quizKeyboard := gotg.Keyboard().
//	    RequestPoll("Create Quiz", true).
//	    BuildReply(gotg.ReplyOptions{Resize: true})
//
// Handling the poll:
//
//	dispatcher.OnMessage(func(ctx *context.Context, u *context.Update) error {
//	    if m, ok := u.EffectiveMessage(); ok {
//	        if poll := m.Poll(); ok {
//	            // Access poll.Poll for the poll data
//	            questions := poll.Poll.Questions
//	        }
//	    }
//	    return nil
//	})
func (k *KeyboardBuilder) RequestPoll(text string, quiz bool) *KeyboardBuilder {
	btn := &tg.KeyboardButtonRequestPoll{
		Text: text,
		Quiz: quiz,
	}
	return k.add(btn)
}

// WebApp adds a button that opens a Telegram Mini App (Web App) when tapped.
//
// Web App buttons are designed for both inline and reply keyboards. When tapped,
// the specified web application opens in a full-screen in-app browser within
// Telegram. The Web App can interact with the bot via the Telegram JavaScript API.
//
// Web Apps are ideal for complex interactive experiences that go beyond what's
// possible with regular Telegram bots, such as games, forms, or dashboards.
//
// The URL must use HTTPS and point to a valid web application that implements
// the Telegram Web App API.
//
// Parameters:
//   - text: The label text to display on the button
//   - url: The HTTPS URL of the web application to open
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    WebApp("Open Game", "https://example.com/game").
//	    Next().
//	    WebApp("Settings", "https://example.com/settings").
//	    Build()
//
// Web App requirements:
// - URL must use HTTPS
// - The web app should include Telegram's Web App script
// - The web app can send data back to the bot via the WebAppData API
//
// For more information on Web Apps, see:
// https://core.telegram.org/bots/webapps
func (k *KeyboardBuilder) WebApp(text, url string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonWebView{
		Text: text,
		URL:  url,
	})
}

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
