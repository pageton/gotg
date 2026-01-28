package main

import (
	"fmt"
	stdlog "log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/log"
	"github.com/pageton/gotg/session"
)

func main() {
	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			InMemory: true,
			Session:  session.SimpleSession(),
			LogConfig: &log.Config{
				MinLevel:  log.LevelDebug,
				Timestamp: true,
				Color:     true,
				Caller:    true,
				FuncName:  true,
			},
		},
	)
	if err != nil {
		stdlog.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher
	dp.AddHandler(handlers.OnCommand("start", start, filters.Private))
	dp.AddHandler(handlers.OnMessage(echo, filters.Message.Text))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		stdlog.Fatalln("failed to start client:", err)
	}
}

func start(u *adapter.Update) error {
	u.Log.Info("received /start", "user", u.GetUserChat().ID)
	_, _ = u.Reply("Hello! I'm a logging demo bot.")
	u.Log.Success("replied to /start")
	return nil
}

func echo(u *adapter.Update) error {
	u.Log.Debug("incoming message", "text", u.Text(), "chat", u.MsgID())
	_, err := u.Reply(u.Text())
	if err != nil {
		u.Log.Error("failed to echo", "err", err)
		return err
	}
	return nil
}
