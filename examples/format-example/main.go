package main

import (
	"fmt"
	"log"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	client, err := gotg.NewClient(
		0,              // APP_ID - Replace with your Telegram App ID from https://my.telegram.org
		"",             // APP_HASH - Replace with your Telegram App Hash from https://my.telegram.org
		gotg.AsBot(""), // BOT_TOKEN - Replace with your bot token from @BotFather
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("formatexample")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	// Add handlers
	dp.AddHandler(handlers.OnCommand("start", start, filters.Private))
	dp.AddHandler(handlers.OnCommand("html", htmlExample, filters.Private))
	dp.AddHandler(handlers.OnCommand("markdown", markdownExample, filters.Private))
	dp.AddHandler(handlers.OnCallbackQuery(filters.CallbackQuery.Prefix("cb_"), buttonCallback))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}

// start demonstrates the ext.Format() helper with rich formatting
func start(u *adapter.Update) error {
	f := u.Format(adapter.HTML)

	// Create a beautifully formatted welcome message
	var text string

	// Header with emoji and bold
	text += "🎉 " + f.Bold("Welcome to the Formatting Demo!") + "\n\n"

	// User info section
	text += f.Bold("👤 Your Information:") + "\n"
	text += f.Italic("Name: ") + f.Code(u.FirstName()+" "+u.LastName()) + "\n"
	text += f.Italic("User ID: ") + f.Code(fmt.Sprintf("%d", u.UserID())) + "\n"
	if u.Username() != "" {
		text += f.Italic("Username: ") + f.Code("@"+u.Username()) + "\n"
	}
	text += "\n"

	// Formatting examples section
	text += f.Underline("✨ Available Formatting Styles:") + "\n\n"

	// Text styles
	text += f.Bold("Bold") + " | "
	text += f.Italic("Italic") + " | "
	text += f.Underline("Underline") + "\n"

	text += f.Strikethrough("Strikethrough") + " | "
	text += f.Spoiler("Spoiler") + " | "
	text += f.Code("Code") + "\n\n"

	// Links
	text += f.Link("Click here for Google", "https://google.com") + "\n"
	text += f.Mention("Mention yourself", u.UserID()) + "\n\n"

	// Code block example
	text += f.Bold("💻 Code Example:") + "\n"
	text += f.PreWithLanguage("fmt.Println('Hello, World!')", "go") + "\n\n"

	// Blockquote
	text += f.Blockquote("This is a blockquote example\nUse it to highlight important text!") + "\n\n"

	// Call to action
	text += f.Italic("Try these commands:") + "\n"
	text += f.Code("/html") + " - HTML formatting demo\n"
	text += f.Code("/markdown") + " - MarkdownV2 formatting demo"

	// Create inline keyboard for quick access
	kbd := gotg.Keyboard().
		Button("📝 HTML", "cb_html").
		Button("📝 Markdown", "cb_markdown").
		Next().
		Button("❓ Help", "cb_help").
		Build()

	_, err := u.Reply(text, &adapter.SendOpts{
		ParseMode:   adapter.HTML,
		ReplyMarkup: kbd,
	})
	return err
}

// htmlExample demonstrates HTML formatting with ext.Format()
func htmlExample(u *adapter.Update) error {
	f := u.Format(adapter.HTML)

	text := "🎨 " + f.Bold("HTML Formatting Examples") + "\n\n"
	text += f.Bold("1. Bold Text") + "\n"
	text += f.Italic("2. Italic Text") + "\n"
	text += f.Underline("3. Underlined Text") + "\n"
	text += f.Strikethrough("4. Strikethrough") + "\n"
	text += f.Spoiler("5. Spoiler (tap to reveal)") + "\n"
	text += f.Code("6. Inline Code") + "\n\n"

	text += f.Bold("7. Pre-formatted Code:") + "\n"
	text += f.PreWithLanguage("const greeting = 'Hello';\nconsole.log(greeting);", "javascript") + "\n\n"

	text += f.Bold("8. Links:") + "\n"
	text += f.Link("Visit Telegram", "https://telegram.org") + " | "
	text += f.Mention("Mention User", u.UserID()) + "\n\n"

	text += f.Bold("9. Blockquote:") + "\n"
	text += f.Blockquote("This is a blockquote.\nPerfect for highlighting quotes or important information.") + "\n\n"

	text += f.Italic("✨ All formatting done with ext.Format()!")

	_, err := u.Reply(text, &adapter.SendOpts{
		ParseMode: adapter.HTML,
	})
	return err
}

// markdownExample demonstrates MarkdownV2 formatting with ext.Format()
func markdownExample(u *adapter.Update) error {
	f := u.Format(adapter.Markdown)

	text := "✨ " + f.Bold("MarkdownV2 Formatting Examples") + "\n\n"
	text += f.Bold("Bold") + " | "
	text += f.Italic("Italic") + " | "
	text += f.Underline("Underline") + "\n\n"

	text += f.Strikethrough("Strikethrough") + " | "
	text += f.Spoiler("Spoiler") + " | "
	text += f.Code("Code") + "\n\n"

	text += f.Bold("Code Block:") + "\n"
	text += f.PreWithLanguage("print('Hello, World!')", "python") + "\n\n"

	text += f.Bold("Links:") + "\n"
	text += f.Link("Telegram", "https://telegram.org") + "\n\n"

	text += f.Blockquote("This is a blockquote in Markdown") + "\n\n"

	text += f.Italic("✨ All formatting done with ext.Format()!")

	_, err := u.Reply(text, &adapter.SendOpts{
		ParseMode: adapter.Markdown,
	})
	return err
}

// buttonCallback handles callback queries from the inline keyboard
func buttonCallback(u *adapter.Update) error {
	if u.CallbackQuery.Data == nil {
		return nil
	}

	data := string(u.CallbackQuery.Data)

	var text string
	var parseMode string
	switch data {
	case "cb_html":
		text = "You chose HTML formatting! 📝\n\nClick below to see the demo:"
		parseMode = adapter.HTML
	case "cb_markdown":
		text = "You chose Markdown formatting! 📝\n\nClick below to see the demo:"
		parseMode = adapter.Markdown
	case "cb_help":
		text = "📚 **Available Commands:**\n\n" +
			"`/start` - Show this welcome message\n" +
			"`/html` - HTML formatting examples\n" +
			"`/markdown` - MarkdownV2 formatting examples"
		parseMode = adapter.Markdown
	default:
		text = "Unknown action"
	}

	// Create keyboard based on selection
	var kbd tg.ReplyMarkupClass
	switch data {
	case "cb_html":
		kbd = gotg.Keyboard().
			Button("🔄 Show Again", "cb_html").
			Button("❓ Help", "cb_help").
			Build()
	case "cb_markdown":
		kbd = gotg.Keyboard().
			Button("🔄 Show Again", "cb_markdown").
			Button("❓ Help", "cb_help").
			Build()
	default:
		kbd = gotg.Keyboard().
			Button("📝 HTML", "cb_html").
			Button("📝 Markdown", "cb_markdown").
			Next().
			Button("❓ Help", "cb_help").
			Build()
	}

	_, _ = u.Answer("✅ Action completed!", &adapter.CallbackOptions{})
	_, err := u.Reply(text, &adapter.SendOpts{
		ParseMode:   parseMode,
		ReplyMarkup: kbd,
	})
	return err
}
