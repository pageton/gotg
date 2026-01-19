package adapter

import (
	"errors"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// SetInlineBotResult invokes method messages.setInlineBotResults#eb5ea206 returning error if any.
// Answer an inline query, for bots only
func (ctx *Context) SetInlineBotResult(request *tg.MessagesSetInlineBotResultsRequest) (bool, error) {
	return ctx.Raw.MessagesSetInlineBotResults(ctx, request)
}

func (ctx *Context) GetInlineBotResults(chatID int64, botUsername string, request *tg.MessagesGetInlineBotResultsRequest) (*tg.MessagesBotResults, error) {
	bot := ctx.PeerStorage.GetPeerByUsername(botUsername)
	if bot.ID == 0 {
		c, err := ctx.ResolveUsername(botUsername)
		if err != nil {
			return nil, err
		}
		switch {
		case c.IsAUser():
			bot = &storage.Peer{
				ID:         c.GetID(),
				AccessHash: c.GetAccessHash(),
			}
		default:
			return nil, errors.New("provided username was invalid for a bot")
		}
	}
	request.Peer, _ = ctx.ResolveInputPeerByID(chatID)
	request.Bot = &tg.InputUser{
		UserID:     bot.ID,
		AccessHash: bot.AccessHash,
	}
	return ctx.Raw.MessagesGetInlineBotResults(ctx, request)
}

// TODO: Implement return helper for inline bot result

// SendInlineBotResult invokes method messages.sendInlineBotResult#7aa11297 returning error if any. Send a result obtained using messages.getInlineBotResults¹.
func (ctx *Context) SendInlineBotResult(chatID int64, request *tg.MessagesSendInlineBotResultRequest) (tg.UpdatesClass, error) {
	if request == nil {
		request = &tg.MessagesSendInlineBotResultRequest{}
	}
	request.RandomID = ctx.generateRandomID()
	if request.Peer == nil {
		request.Peer, _ = ctx.ResolveInputPeerByID(chatID)
	}
	return ctx.Raw.MessagesSendInlineBotResult(ctx, request)
}
