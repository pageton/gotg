package gotgproto

import (
	"context"
	"errors"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type RecaptchaSolver interface {
	SolveRecaptcha(packageID, action, key string) (string, error)
}

func parseRecaptcha(text string) (string, string) {
	start := strings.Index(text, "RECAPTCHA_CHECK_")
	if start == -1 {
		return "", ""
	}
	start += len("RECAPTCHA_CHECK_")
	if start >= len(text) {
		return "", ""
	}
	payload := text[start:]
	sep := strings.Index(payload, "__")
	if sep == -1 {
		return "", ""
	}
	action := payload[:sep+1]
	key := payload[sep+2:]
	if action == "" || key == "" {
		return "", ""
	}
	return action, key
}

type FlowClient struct {
	auth.FlowClient
	api     *tg.Client
	appID   int
	apiHash string
	params  tg.JSONValueClass
	solver  RecaptchaSolver
}

func (c FlowClient) SendCode(ctx context.Context, phone string, opts auth.SendCodeOptions) (tg.AuthSentCodeClass, error) {
	sentCode, err := c.FlowClient.SendCode(ctx, phone, opts)
	if err == nil {
		return sentCode, err
	}
	action, key := parseRecaptcha(err.Error())
	if action == "" || key == "" {
		return nil, err
	}
	packageID := ""
	if obj, ok := c.params.(*tg.JSONObject); ok {
		for _, item := range obj.Value {
			if item.Key == "package_id" || item.Key == "bundleId" {
				if v, ok := item.Value.(*tg.JSONString); ok {
					packageID = v.Value
				}
				break
			}
		}
	}
	token, err := c.solver.SolveRecaptcha(packageID, action, key)
	if err != nil {
		return nil, err
	}

	var settings tg.CodeSettings
	if opts.AllowAppHash {
		settings.SetAllowAppHash(true)
	}
	if opts.AllowFlashCall {
		settings.SetAllowFlashcall(true)
	}
	if opts.CurrentNumber {
		settings.SetCurrentNumber(true)
	}

	req := &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       c.appID,
		APIHash:     c.apiHash,
		Settings:    settings,
	}

	var box tg.AuthSentCodeBox
	if err := c.api.Invoker().Invoke(ctx, &tg.InvokeWithReCaptchaRequest{
		Token: token,
		Query: req,
	}, &box); err != nil {
		return nil, err
	}
	if box.SentCode == nil {
		return nil, errors.New("send code returned empty data")
	}
	return box.SentCode, nil
}
