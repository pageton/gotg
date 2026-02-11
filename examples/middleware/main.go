package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"golang.org/x/time/rate"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
)

func main() {
	// Type of client to login to, same as in https://github.com/pageton/gotg/blob/beta/examples/echo-bot/memory_session/main.go#L17
	clientType := gotg.AsBot("BOT_TOKEN_HERE")

	// Initializing flood waiter, which will wait for stated duration if "FLOOD_WAIT" error occurred
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Printf("Waiting for flood, dur: %d\n", wait.Duration)
	})
	// Initializing ratelimiter, which will allow at most 30 requests to Telegram in 100ms
	ratelimiter := ratelimit.New(rate.Every(time.Millisecond*100), 30)

	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		clientType,
		&gotg.ClientOpts{
			InMemory:    true,
			Session:     session.SimpleSession(),
			Middlewares: []telegram.Middleware{waiter, ratelimiter},
			RunMiddleware: func(origRun func(ctx context.Context, f func(ctx context.Context) error) (err error), ctx context.Context, f func(ctx context.Context) (err error)) (err error) {
				return origRun(ctx, func(ctx context.Context) error {
					return waiter.Run(ctx, f)
				})
			},
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	clientDispatcher := client.Dispatcher

	clientDispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, echo), 1)

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
}

func echo(ctx *adapter.Context, update *adapter.Update) error {
	_, err := update.Reply(update.EffectiveMessage.Text, nil)
	return err
}
