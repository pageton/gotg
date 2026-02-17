package adapter

import (
	"github.com/pageton/gotg/i18n"
	"golang.org/x/text/language"
)

type Translator interface {
	Get(userID int64, key string, args ...any) string
	GetCtx(userID int64, key string, ctx *i18n.Args) string
	SetLang(userID int64, lang any)
	GetLang(userID int64) any
}

var globalTranslator Translator

func SetTranslator(t Translator) {
	globalTranslator = t
}

func updateTImpl(u *Update, key string, args ...any) string {
	if globalTranslator == nil {
		return key
	}

	if len(args) > 0 {
		if argsCtx, ok := args[len(args)-1].(*i18n.Args); ok {
			return globalTranslator.GetCtx(u.UserID(), key, argsCtx)
		}
	}

	return globalTranslator.Get(u.UserID(), key, args...)
}

func updateSetLangImpl(u *Update, lang any) {
	if globalTranslator == nil {
		return
	}
	globalTranslator.SetLang(u.UserID(), lang)
}

func updateGetLangImpl(u *Update) any {
	if globalTranslator == nil {
		return language.English
	}
	return globalTranslator.GetLang(u.UserID())
}
