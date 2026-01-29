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
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		// Bot token from @BotFather
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("business-bot.db")),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	dp := client.Dispatcher

	dp.AddHandler(handlers.OnBusinessConnection(onBusinessConnection, filters.BusinessConn.Enabled))
	dp.AddHandler(handlers.OnBusinessMessage(onBusinessMessage, filters.Message.Text))
	dp.AddHandler(handlers.OnBusinessEditedMessage(onBusinessEditedMessage))
	dp.AddHandler(handlers.OnBusinessDeletedMessage(onBusinessDeletedMessage))
	dp.AddHandler(handlers.OnBusinessCallbackQuery(onBusinessCallbackQuery))

	log.Printf("Business bot @%s started.", client.Self.Username)

	client.Idle()
}

func onBusinessConnection(u *adapter.Update) error {
	conn := u.BusinessConnection.Connection
	log.Printf("[BusinessConnection] User %d connected (ID: %s, DC: %d)",
		conn.UserID, conn.ConnectionID, conn.DCID)
	return nil
}

func onBusinessMessage(u *adapter.Update) error {
	_, err := u.Reply(u.Text(), &adapter.SendOpts{
		Entities:  u.EffectiveMessage.Entities,
		ParseMode: adapter.ModeNone,
	})
	return err
}

func onBusinessEditedMessage(u *adapter.Update) error {
	log.Printf("[BusinessEditedMessage] Connection %s: message %d edited",
		u.ConnectionID(), u.MsgID())
	return nil
}

func onBusinessDeletedMessage(u *adapter.Update) error {
	log.Printf("[BusinessDeletedMessage] Connection %s: %d messages deleted",
		u.ConnectionID(), len(u.BusinessDeletedMessages.Messages))
	return nil
}

func onBusinessCallbackQuery(u *adapter.Update) error {
	log.Printf("[BusinessCallbackQuery] Connection %s: data=%s",
		u.ConnectionID(), string(u.BusinessCallbackQuery.Data))
	return nil
}
