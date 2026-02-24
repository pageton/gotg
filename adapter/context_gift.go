package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// GetStarGifts retrieves the list of available star gifts.
func (ctx *Context) GetStarGifts(hash int) (tg.PaymentsStarGiftsClass, error) {
	return functions.GetStarGifts(ctx.Context, ctx.Raw, hash)
}

// GetResaleStarGifts retrieves the list of resale star gifts.
func (ctx *Context) GetResaleStarGifts(req *tg.PaymentsGetResaleStarGiftsRequest) (*tg.PaymentsResaleStarGifts, error) {
	return functions.GetResaleStarGifts(ctx.Context, ctx.Raw, req)
}

// BuyResaleStarGift purchases a resale star gift by slug for a recipient.
// toID can be tg.InputPeerClass, int64 (userID), or string (username).
func (ctx *Context) BuyResaleStarGift(slug string, toID any) (tg.PaymentsPaymentResultClass, error) {
	return functions.BuyResaleStarGift(ctx.Context, ctx.Raw, ctx.PeerStorage, slug, toID)
}
