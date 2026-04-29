package broadcast_test

import (
	"context"
	"fmt"
	"time"

	"github.com/pageton/gotg/broadcast"
)

// ExampleBroadcaster demonstrates how to use the Broadcaster with a GoTG bot.
// In production, replace the MakeSendFunc with broadcast.MakeTextSendFunc
// for real message sending.
func ExampleBroadcaster() {
	// Define targets — callers resolve identifiers to chat IDs before constructing.
	targets := []broadcast.PeerTarget{
		{ChatID: 100},
		{ChatID: 200},
		{ChatID: 300},
		{ChatID: 400},
	}

	cfg := broadcast.Config{
		Workers:      5,
		MaxRetries:   3,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
	}

	b := broadcast.New(targets, cfg)

	// In a real bot handler (func(u *adapter.Update) error):
	//
	//   makeSend := broadcast.MakeTextSendFunc(u.Ctx, broadcast.TextSendConfig{
	//       Text: "Hello from the broadcast system!",
	//   })
	//   result := b.Run(u.Ctx, makeSend)
	//
	// For this example, use a mock sender:
	result := b.Run(context.Background(), func(chatID int64) broadcast.SendFunc {
		return func(attempt int) error {
			// In production: this calls ctx.SendMessage internally.
			return nil
		}
	})

	fmt.Printf("sent=%d failed=%d skipped=%d total=%d\n",
		result.Sent(), result.Failed(), result.Skipped(), result.Total())

	// Output:
	// sent=4 failed=0 skipped=0 total=4
}
