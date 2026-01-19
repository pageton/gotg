package adapter

import (
	"github.com/gotd/td/tg"
)

// AnswerCallback invokes method messages.setBotCallbackAnswer#d58f130a returning error if any. Set the callback answer to a user button press
func (ctx *Context) AnswerCallback(request *tg.MessagesSetBotCallbackAnswerRequest) (bool, error) {
	if request == nil {
		request = &tg.MessagesSetBotCallbackAnswerRequest{}
	}
	return ctx.Raw.MessagesSetBotCallbackAnswer(ctx, request)
}

// Data returns the callback data as a string for callback queries.
// Returns empty string if not a callback query or data is nil.
func (u *Update) Data() string {
	if u.CallbackQuery != nil && u.CallbackQuery.Data != nil {
		return string(u.CallbackQuery.Data)
	}
	return ""
}
