package dispatcher

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/i18n"
	"golang.org/x/text/language"
)

// I18nConfig configures the i18n middleware
type I18nConfig struct {
	// Translator is the i18n translator instance
	Translator *i18n.Translator
	// DefaultLanguage is the default language to use
	DefaultLanguage language.Tag
	// DetectFromTelegram automatically detects language from Telegram user data
	DetectFromTelegram bool
}

// I18nMiddleware creates a middleware handler that adds i18n support to all updates.
// The middleware runs in group 0 (before all other handlers) and initializes the translator.
//
// Example:
//
//	translator := i18n.NewTranslator(&i18n.LocaleConfig{
//	    DefaultLang: language.English,
//	    Session:     peerStorage,
//	})
//	dp.AddHandlerToGroup(I18nMiddleware(&I18nConfig{
//	    Translator:         translator,
//	    DetectFromTelegram: true,
//	}), 0)
func I18nMiddleware(config *I18nConfig) Handler {
	return &i18nMiddlewareHandler{
		config: config,
	}
}

type i18nMiddlewareHandler struct {
	config *I18nConfig
}

// translatorAdapter adapts i18n.Translator to ext.Translator interface
type translatorAdapter struct {
	*i18n.Translator
}

func (a *translatorAdapter) GetCtx(userID int64, key string, ctx *i18n.Args) string {
	return a.Translator.GetCtx(userID, key, ctx)
}

func (a *translatorAdapter) SetLang(userID int64, lang any) {
	if l, ok := lang.(language.Tag); ok {
		a.Translator.SetLang(userID, l)
	}
}

func (a *translatorAdapter) GetLang(userID int64) any {
	return a.Translator.GetLang(userID)
}

// CheckUpdate initializes the translator for each update and passes it through.
// This middleware does not consume the update - it always returns ContinueGroups.
func (m *i18nMiddlewareHandler) CheckUpdate(ctx *adapter.Context, update *adapter.Update) error {
	if m.config == nil || m.config.Translator == nil {
		return ContinueGroups
	}

	// Set the global translator for this update using adapter
	i18nAdapter := &translatorAdapter{Translator: m.config.Translator}
	adapter.SetTranslator(i18nAdapter)

	// Auto-detect language from Telegram user if enabled
	if m.config.DetectFromTelegram {
		tgUser := update.EffectiveUser()
		if tgUser != nil {
			// Initialize user language from Telegram if not already set
			m.config.Translator.InitFromUser(&tgUser.User)
		}
	}

	// Continue processing the update
	return ContinueGroups
}
