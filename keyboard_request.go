package gotg

import (
	"github.com/gotd/td/tg"
)

// Game adds an HTML5 game button to the keyboard.
//
// Game buttons launch a Telegram HTML5 game when tapped. The bot must have
// a game registered via @BotFather. When pressed, Telegram opens the game
// interface for the user.
//
// Parameters:
//   - text: The label text to display on the button
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Game("Play Now").
//	    Build()
func (k *KeyboardBuilder) Game(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonGame{Text: text})
}

// Buy adds a payment button to the keyboard.
//
// Buy buttons are used with Telegram's payment API. When tapped, the user is
// prompted to complete a payment flow. The bot must have a payment provider
// configured via @BotFather.
//
// Parameters:
//   - text: The label text to display on the button
//
// Returns the builder for method chaining.
//
// Example:
//
//	keyboard := gotg.Keyboard().
//	    Buy("Pay $9.99").
//	    Build()
func (k *KeyboardBuilder) Buy(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonBuy{Text: text})
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
