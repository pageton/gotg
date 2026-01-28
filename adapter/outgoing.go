package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

const (
	ActionSend   = "send"
	ActionEdit   = "edit"
	ActionDelete = "delete"
)

const (
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// FakeOutgoingUpdate is a synthetic update emitted when the client
// sends, edits, or deletes a message. Raw MTProto does not push
// outgoing message updates; this struct fills that gap.
type FakeOutgoingUpdate struct {
	Action    string
	Status    string
	Message   *types.Message
	MessageID int
	ChatID    int64
	Peer      tg.InputPeerClass
	Error     error
}
