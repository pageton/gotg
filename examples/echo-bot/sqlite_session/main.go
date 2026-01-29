package main

import (
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		// ClientType, as we defined above
		gotg.AsBot("BOT_TOKEN_HERE"),
		// Optional parameters of client
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("echobot")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	// Command Handler for /start
	dp.AddHandler(handlers.OnCommand("start", start, filters.Private))
	// Callback Query Handler with prefix filter for recieving specific query
	dp.AddHandler(handlers.OnCallbackQuery(filters.CallbackQuery.Prefix("cb_"), buttonCallback))
	// This Message Handler will call our echo function on new messages
	dp.AddHandler(handlers.OnMessage(echo, filters.Message.Text))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
}

// callback function for /start command
func start(u *adapter.Update) error {
	kbd := gotg.Keyboard().
		URL("gotd/td", "https://github.com/gotd/td").
		URL("GoTG", "https://github.com/pageton/gotg").
		Next().
		Button("Click Here", "cb_pressed").
		Build()
	_, _ = u.Reply(fmt.Sprintf("Hello %s, I am @%s and will repeat all your messages.\nI was made using gotd and gotg.", u.Mention(), u.Self.Username), &adapter.SendOpts{ReplyMarkup: kbd})
	// End dispatcher groups so that bot doesn't echo /start command usage
	return dispatcher.EndGroups
}

func buttonCallback(u *adapter.Update) error {
	_, _ = u.Answer("This is an example bot!", &adapter.CallbackOptions{Alert: true})
	return nil
}

func echo(u *adapter.Update) error {
	_, err := u.Reply(u.Text())
	return err
}
