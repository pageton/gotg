package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// TransferStarGift transfers a star gift to a chat.
func (ctx *Context) TransferStarGift(chatID int64, starGift tg.InputSavedStarGiftClass) (tg.UpdatesClass, error) {
	return functions.TransferStarGift(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, starGift)
}

// ExportInvoice exports an invoice for payment providers.
func (ctx *Context) ExportInvoice(inputMedia tg.InputMediaClass) (*tg.PaymentsExportedInvoice, error) {
	return functions.ExportInvoice(ctx.Context, ctx.Raw, inputMedia)
}

// SetPreCheckoutResults responds to a pre-checkout query.
func (ctx *Context) SetPreCheckoutResults(success bool, queryID int64, err string) (bool, error) {
	return functions.SetPreCheckoutResults(ctx.Context, ctx.Raw, success, queryID, err)
}

// PinMessage pins a message in a chat.
func (ctx *Context) PinMessage(chatID int64, messageID int) (tg.UpdatesClass, error) {
	return functions.PinMessage(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, messageID)
}

// UnPinMessage unpins a specific message in a chat.
func (ctx *Context) UnPinMessage(chatID int64, messageID int) error {
	return functions.UnPinMessage(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, messageID)
}

// UnPinAllMessages unpins all messages in a chat.
func (ctx *Context) UnPinAllMessages(chatID int64) error {
	return functions.UnPinAllMessages(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID)
}

// GetChatInviteLink generates an invite link for a chat.
func (ctx *Context) GetChatInviteLink(chatID int64, req ...*tg.MessagesExportChatInviteRequest) (tg.ExportedChatInviteClass, error) {
	return functions.GetChatInviteLink(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, req...)
}
