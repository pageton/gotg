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
// Returns an empty string if the update is not a callback query or data is nil.
//
// Example:
//
//	data := u.Data()
//	if data != "" {
//	    fmt.Printf("Callback data: %s\n", data)
//	}
func (u *Update) Data() string {
	if u.CallbackQuery != nil && u.CallbackQuery.Data != nil {
		return string(u.CallbackQuery.Data)
	}
	return ""
}
