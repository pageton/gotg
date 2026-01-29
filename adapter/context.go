package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/types"
)

// IsBot returns true if the current session belongs to a bot account.
func (ctx *Context) IsBot() bool {
	return ctx.Self != nil && ctx.Self.Bot
}

func (ctx *Context) emitOutgoing(action, status string, msg *types.Message, msgID int, chatID int64, peer tg.InputPeerClass, err error) {
	if ctx.OnOutgoing == nil {
		return
	}
	ctx.OnOutgoing(&FakeOutgoingUpdate{
		Action:    action,
		Status:    status,
		Message:   msg,
		MessageID: msgID,
		ChatID:    chatID,
		Peer:      peer,
		Error:     err,
	})
}

// generateRandomID generates a random int64 for use in Telegram API calls.
// Random IDs are required by Telegram for duplicate request prevention.
// Uses thread-safe shared random source to prevent memory leaks (Issue #112).
func (ctx *Context) generateRandomID() int64 {
	return functions.GenerateRandomID()
}
