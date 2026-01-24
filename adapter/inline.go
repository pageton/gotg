package adapter

import (
	"errors"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// SetInlineBotResult answers an inline query with results.
// For bots only. This is the method to respond to inline queries.
//
// Parameters:
//   - request: The inline bot results request containing query ID and results
//
// Returns true if successful, or an error.
//
// Example:
//
//	results := &tg.MessagesSetInlineBotResultsRequest{
//	    QueryID: queryID,
//	    Results: []tg.InputBotInlineResultClass{...},
//	}
//	success, err := ctx.SetInlineBotResult(results)
func (ctx *Context) SetInlineBotResult(request *tg.MessagesSetInlineBotResultsRequest) (bool, error) {
	return ctx.Raw.MessagesSetInlineBotResults(ctx, request)
}

// GetInlineBotResults fetches inline results from a bot.
//
// Parameters:
//   - chatID: The chat ID where the query originates
//   - botUsername: The username of the bot to query (with or without @ prefix)
//   - request: The request containing query parameters
//
// Returns the bot results or an error.
func (ctx *Context) GetInlineBotResults(chatID int64, botUsername string, request *tg.MessagesGetInlineBotResultsRequest) (*tg.MessagesBotResults, error) {
	bot := ctx.PeerStorage.GetPeerByUsername(botUsername)
	if bot.ID == 0 {
		c, err := ctx.ResolveUsername(botUsername)
		if err != nil {
			return nil, err
		}
		switch {
		case c.IsUser():
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

// SendInlineBotResult sends an inline bot result to a chat.
// Used to send a result obtained from GetInlineBotResults.
//
// Parameters:
//   - chatID: The chat ID to send the result to
//   - request: The send request containing the result to send
//
// Returns updates confirming the action or an error.
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
