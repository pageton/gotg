package main

import (
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
)

// This example shows how to use a custom storage adapter (JSON file).
// See json_adapter.go for the full implementation.
//
// You can use the same pattern for any database:
//   MongoDB, BoltDB, BadgerDB, DynamoDB, etc.
//
// Built-in adapters:
//   session.SqlSession(sqlite.Open("bot.db"))    — SQLite via GORM (modernc or mattn)
//   session.SqlSession(postgres.Open(dsn))        — PostgreSQL via GORM
//   session.SqlSession(mysql.Open(dsn))           — MySQL via GORM
//   session.WithAdapter(redisdb.New(redisClient)) — Redis
//   session.WithAdapter(sqlcdb.New(sqlDB))        — raw SQL (sqlc-style)
//   session.WithAdapter(myAdapter)                — any custom adapter

func main() {
	jsonAdapter, err := NewJsonAdapter("bot_session.json")
	if err != nil {
		log.Fatalln("failed to create json adapter:", err)
	}

	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			Session: session.WithAdapter(jsonAdapter),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher
	dp.AddHandler(handlers.OnCommand("start", func(u *adapter.Update) error {
		_, err := u.Reply(fmt.Sprintf("Hello %s! Session stored in JSON.", u.Mention()))
		return err
	}, filters.Private))

	dp.AddHandler(handlers.OnMessage(func(u *adapter.Update) error {
		_, err := u.Reply(u.Text())
		return err
	}, filters.Message.Text))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)
	client.Idle()
}
