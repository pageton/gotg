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
)

func main() {
	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			InMemory:     true,
			SendOutgoing: true,
			Session:      session.SimpleSession(),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	dp.AddHandler(handlers.OnCommand("start", func(u *adapter.Update) error {
		_, _ = u.Reply("Hello! Send me a message and I'll echo it back.")
		return dispatcher.EndGroups
	}))

	dp.AddHandler(handlers.OnMessage(echo, filters.Message.Text))

	dp.AddHandler(handlers.OnOutgoing(outgoing))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
}

func echo(u *adapter.Update) error {
	_, err := u.Reply(u.Text())
	return err
}

func outgoing(u *adapter.Update) error {
	ou := u.EffectiveOutgoing
	switch ou.Status {
	case adapter.StatusSucceeded:
		u.Log.Successf("%s succeeded: msg %d in chat %d", ou.Action, ou.MessageID, ou.ChatID)
	case adapter.StatusFailed:
		u.Log.Errorf("%s failed: %v", ou.Action, ou.Error)
	}
	return nil
}
