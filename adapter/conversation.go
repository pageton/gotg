// Package adapter provides conversation management functionality for the gotg framework.
// It enables handling multi-step conversations with users through state tracking
// and message filtering capabilities.
package adapter

import (
	"encoding/json"
	"time"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/types"
)

// ConvOpts defines optional configuration parameters for starting a conversation.
// It provides control over message formatting, reply behavior, media attachment,
// and conversation timeout settings.
type ConvOpts struct {
	// Reply indicates whether to reply to the user's message. When true,
	// the sent message will reply to the original triggering message.
	Reply bool
	// ReplyMarkup provides inline or reply keyboard markup for the message.
	// This enables interactive buttons and custom keyboards.
	ReplyMarkup tg.ReplyMarkupClass
	// ParseMode specifies the text formatting (e.g., "html", "markdown").
	// Controls how the message text should be parsed and formatted.
	ParseMode string
	// Media attaches a media file (photo, video, document, etc.) to the message.
	// When set, the message becomes a media message rather than text-only.
	Media tg.InputMediaClass
	// Caption is the text caption for media messages. Ignored when Media is nil.
	Caption string
	// Timeout sets the duration before the conversation state automatically expires.
	// A zero value means no timeout (state persists until manually cleared).
	Timeout time.Duration
	// Filter specifies which incoming updates should be accepted during the conversation.
	// Only updates matching the filter will be processed and returned to the handler.
	Filter conv.Filter
}

// StartConv initiates a new conversation step with the user.
// It stores the conversation state and sends an optional message to the user.
// The conversation is identified by the current chat and user ID from the update.
//
// Parameters:
//   - step: A unique identifier for this conversation step (e.g., "awaiting_name")
//   - text: The message text to send to the user
//   - opts: Optional configuration for message sending and conversation behavior
//
// Returns the sent message and any error encountered.
func (u *Update) StartConv(step string, text string, opts ...*ConvOpts) (*types.Message, error) {
	return u.startConv(step, text, nil, opts...)
}

// StartConvWithData initiates a new conversation step with custom data payload.
// Unlike StartConv, this method stores additional data that can be retrieved
// when the user responds, enabling stateful multi-step conversations.
//
// Parameters:
//   - step: A unique identifier for this conversation step
//   - text: The message text to send to the user
//   - data: Arbitrary data to store with the conversation state (will be JSON marshaled)
//   - opts: Optional configuration for message sending and conversation behavior
//
// Returns the sent message and any error encountered.
func (u *Update) StartConvWithData(step string, text string, data map[string]any, opts ...*ConvOpts) (*types.Message, error) {
	return u.startConv(step, text, data, opts...)
}

// startConv is the internal implementation for starting a conversation.
// It handles setting the conversation state with optional timeout and filter,
// then sends the initial message to the user.
func (u *Update) startConv(step string, text string, data map[string]any, opts ...*ConvOpts) (*types.Message, error) {
	key := conv.Key{ChatID: u.ChatID(), UserID: u.UserID()}

	var opt *ConvOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	var payload []byte
	if data != nil {
		payload, _ = json.Marshal(data)
	}

	var timeout time.Duration
	var filter conv.Filter
	if opt != nil {
		timeout = opt.Timeout
		filter = opt.Filter
	}

	if err := u.Ctx.Conv.SetStateWithOpts(key, step, payload, timeout, filter); err != nil {
		return nil, err
	}

	return u.sendConvMessage(text, opt)
}

// sendConvMessage sends a message as part of a conversation.
// It handles both text messages and media messages based on the provided options.
// The message will reply to the original update if the Reply option is set.
func (u *Update) sendConvMessage(text string, opt *ConvOpts) (*types.Message, error) {
	if opt == nil {
		return u.SendMessage(0, text, nil)
	}

	if opt.Media != nil {
		mediaOpts := &ReplyMediaOpts{
			Markup:    opt.ReplyMarkup,
			ParseMode: opt.ParseMode,
			Caption:   opt.Caption,
		}
		if opt.Reply {
			mediaOpts.ReplyMessageID = u.MsgID()
		}
		return u.SendMedia(opt.Media, text, mediaOpts)
	}

	replyOpts := &ReplyOpts{
		Markup:    opt.ReplyMarkup,
		ParseMode: opt.ParseMode,
	}
	if opt.Reply {
		replyOpts.ReplyMessageID = u.MsgID()
	}
	return u.SendMessage(0, text, replyOpts)
}

// EndConv terminates the active conversation for the current chat and user.
// It clears both the conversation state and any associated filter.
//
// Returns:
//   - cleared: True if the conversation state was successfully cleared
//   - existed: True if a conversation was active (false if none existed)
//   - err: Any error encountered during the operation
func (u *Update) EndConv() (bool, bool, error) {
	key := conv.Key{ChatID: u.ChatID(), UserID: u.UserID()}

	state, err := u.Ctx.Conv.GetState(key)
	if err != nil {
		return false, false, err
	}
	if state == nil {
		return false, false, nil
	}

	err = u.Ctx.Conv.ClearState(key)
	u.Ctx.Conv.ClearFilter(key)
	return err == nil, true, err
}

// CancelConv is an alias for EndConv that provides more semantic clarity
// when a conversation is being cancelled rather than naturally ended.
// Both methods behave identically.
func (u *Update) CancelConv() (bool, bool, error) {
	return u.EndConv()
}
