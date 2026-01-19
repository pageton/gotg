package adapter

import (
	"reflect"

	"github.com/pageton/gotg/i18n"
	"golang.org/x/text/language"
)

// Translator interface defines the methods needed for i18n
type Translator interface {
	Get(userID int64, key string, args ...any) string
	GetCtx(userID int64, key string, ctx *i18n.Args) string
	SetLang(userID int64, lang any)
	GetLang(userID int64) any
}

// Global translator set by middleware
var globalTranslator Translator

// SetTranslator sets the global translator (called by middleware)
func SetTranslator(t Translator) {
	globalTranslator = t
}

// updateTImpl implements the T method
// Checks if args contains *i18n.Args and uses GetCtx if so
func updateTImpl(u *Update, key string, args ...any) string {
	if globalTranslator == nil {
		return key
	}

	// Check if last arg is *i18n.Args for context-based translation
	if len(args) > 0 {
		if lastArg := args[len(args)-1]; lastArg != nil {
			if argType := reflect.TypeOf(lastArg); argType != nil {
				// Check if it's *i18n.Args by checking package and type name
				if argType.Kind() == reflect.Pointer {
					if elemType := argType.Elem(); elemType != nil {
						pkgPath := elemType.PkgPath()
						typeName := elemType.Name()
						if pkgPath == "github.com/pageton/gotg/i18n" && typeName == "Args" {
							// This is *i18n.Args, use GetCtx
							if argsPtr, ok := lastArg.(*i18n.Args); ok {
								return globalTranslator.GetCtx(u.UserID(), key, argsPtr)
							}
						}
					}
				}
			}
		}
	}

	// Use Get with positional args
	return globalTranslator.Get(u.UserID(), key, args...)
}

// updateSetLangImpl implements the SetLang method
func updateSetLangImpl(u *Update, lang any) {
	if globalTranslator == nil {
		return
	}
	// Convert language.Tag to any if needed
	if l, ok := lang.(language.Tag); ok {
		globalTranslator.SetLang(u.UserID(), l)
	} else {
		globalTranslator.SetLang(u.UserID(), lang)
	}
}

// updateGetLangImpl implements the GetLang method
func updateGetLangImpl(u *Update) any {
	if globalTranslator == nil {
		return language.English
	}
	return globalTranslator.GetLang(u.UserID())
}
