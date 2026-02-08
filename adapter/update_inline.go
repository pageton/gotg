package adapter

import (
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

// AnswerInlineQuery answers the inline query with results.
// results: The slice of inline results to show.
// opts: Optional *types.AnswerOpts for cacheTime, isPersonal, nextOffset, etc.
//
// Example:
//
//	results := []tg.InputBotInlineResultClass{...}
//	u.AnswerInlineQuery(results, nil)
//	u.AnswerInlineQuery(results, &types.AnswerOpts{CacheTime: 60, IsPersonal: true})
func (u *Update) AnswerInlineQuery(results []tg.InputBotInlineResultClass, opts *types.AnswerOpts) (bool, error) {
	if u.InlineQuery == nil {
		return false, fmt.Errorf("no inline query in this update")
	}
	return u.InlineQuery.Answer(results, opts)
}

// AnswerInlineQueryWithGallery answers the inline query with results in gallery format.
// Use this for photo/video results that should display as a grid.
func (u *Update) AnswerInlineQueryWithGallery(results []tg.InputBotInlineResultClass, opts *types.AnswerOpts) (bool, error) {
	if u.InlineQuery == nil {
		return false, fmt.Errorf("no inline query in this update")
	}
	return u.InlineQuery.AnswerWithGallery(results, opts)
}
