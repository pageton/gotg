package main

import (
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("bot.db")),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use the client's built-in dispatcher
	dp := client.Dispatcher

	// Add command handlers to group 0
	// dp.AddHandler(handlers.OnCommand("start", startHandler), 0)
	dp.AddHandlers(handlers.OnCommand("start", startHandler), handlers.OnCommand("echo", echoHandler), handlers.OnCommand("help", helpHandler))
	// dp.AddHandlerToGroup(handlers.OnCommand("help", helpHandler), 0)

	// Add message handler to group 1
	dp.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, handlers.ToCallbackResponse(echoHandler)), 1)

	log.Printf("Bot @%s started. Send /start to begin.", client.Self.Username)

	client.Idle()
}

func startHandler(u *adapter.Update) error {
	log.Printf("[Start] User %d sent /start command", u.UserID())
	u.Reply("Welcome! Use /help to see available commands.")
	return nil
}

func helpHandler(u *adapter.Update) error {
	log.Printf("[Help] User %d sent /help command", u.UserID())
	u.Reply(`
Available commands:
/start - Start the bot
/help - Show this help message
/echo - Echo back your message
	`)
	return nil
}

func echoHandler(u *adapter.Update) error {
	msg := u.EffectiveMessage
	if msg == nil {
		return nil
	}
	log.Printf("[Echo] User %d sent: %s", u.UserID(), msg.Text)
	u.Reply("Echo: " + msg.Text)
	return nil
}
