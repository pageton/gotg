package adapter

import (
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conversation"
	"github.com/pageton/gotg/types"
)

// AskOption customizes ctx.Ask behavior.
type AskOption func(*askConfig)

type askConfig struct {
	Timeout time.Duration
	Filter  conversation.Filter
	Step    string
	Payload []byte
	Request *tg.MessagesSendMessageRequest
}

// AskWithTimeout overrides the wait timeout.
func AskWithTimeout(d time.Duration) AskOption {
	return func(cfg *askConfig) {
		cfg.Timeout = d
	}
}

// AskWithFilter applies a custom filter.
func AskWithFilter(f conversation.Filter) AskOption {
	return func(cfg *askConfig) {
		cfg.Filter = f
	}
}

// AskWithStep tags the persisted state with a custom step label.
func AskWithStep(step string) AskOption {
	return func(cfg *askConfig) {
		cfg.Step = step
	}
}

// AskWithPayload stores opaque metadata alongside the conversation state.
func AskWithPayload(payload []byte) AskOption {
	return func(cfg *askConfig) {
		cfg.Payload = payload
	}
}

// AskWithRequest allows providing a fully customized send request.
func AskWithRequest(req *tg.MessagesSendMessageRequest) AskOption {
	return func(cfg *askConfig) {
		cfg.Request = req
	}
}

// Ask sends a prompt to chatID and waits for the user's next reply matching the filter.
func (ctx *Context) Ask(chatID, userID int64, prompt string, opts ...AskOption) (*types.Message, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if ctx.Conversations == nil {
		return nil, fmt.Errorf("conversation manager is not enabled")
	}
	if chatID == 0 || userID == 0 {
		return nil, fmt.Errorf("ask requires chatID and userID")
	}

	cfg := askConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.Request == nil {
		cfg.Request = &tg.MessagesSendMessageRequest{Message: prompt}
	} else if cfg.Request.Message == "" {
		cfg.Request.Message = prompt
	}
	if cfg.Request.RandomID == 0 {
		cfg.Request.RandomID = ctx.generateRandomID()
	}
	sent, err := ctx.SendMessage(chatID, cfg.Request)
	if err != nil {
		return nil, err
	}
	key := conversation.Key{ChatID: chatID, UserID: userID}
	resp, err := ctx.Conversations.WaitResponse(ctx.Context, key, conversation.Options{
		Timeout: cfg.Timeout,
		Filter:  cfg.Filter,
		Step:    cfg.Step,
		Payload: cfg.Payload,
	})
	if err != nil {
		return nil, err
	}
	resp.ReplyToMessage = sent
	return resp, nil
}

// Ask prompts the user originating this update.
func (u *Update) Ask(prompt string, opts ...AskOption) (*types.Message, error) {
	if u == nil || u.Ctx == nil {
		return nil, fmt.Errorf("update context missing")
	}
	chatID := u.ChatID()
	userID := u.UserID()
	if chatID == 0 || userID == 0 {
		return nil, fmt.Errorf("unable to resolve chat/user for ask")
	}
	return u.Ctx.Ask(chatID, userID, prompt, opts...)
}
