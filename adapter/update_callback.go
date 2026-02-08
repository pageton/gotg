package adapter

import (
	"fmt"

	"github.com/gotd/td/tg"
)

// Answer answers the callback query (works for both regular and inline message callbacks).
// text: The notification text (use empty string for silent).
// opts: Optional *CallbackOptions for alert, cacheTime, url.
//
// Example:
//
//	u.Answer("Done!", nil)
//	u.Answer("Error!", &CallbackOptions{Alert: true})
func (u *Update) Answer(text string, opts ...*CallbackOptions) (bool, error) {
	var queryID int64

	switch {
	case u.CallbackQuery != nil:
		queryID = u.CallbackQuery.QueryID
	case u.InlineCallbackQuery != nil:
		queryID = u.InlineCallbackQuery.QueryID
	default:
		return false, fmt.Errorf("no callback query in this update")
	}

	alert := false
	cacheTime := 0
	url := ""

	if len(opts) > 0 && opts[0] != nil {
		alert = opts[0].Alert
		cacheTime = opts[0].CacheTime
		url = opts[0].URL
	}

	return u.Ctx.Raw.MessagesSetBotCallbackAnswer(u.Ctx, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   queryID,
		Message:   text,
		Alert:     alert,
		CacheTime: cacheTime,
		URL:       url,
	})
}
