package dispatcher

import (
	"slices"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conv"
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

type ConvOpts struct {
	Manager        *conv.Manager
	CancelKeywords []string
	CancelReply    bool
	CancelOpts     *adapter.SendOpts
	CancelText     string
}

func Conv(opts ConvOpts) Handler {
	return &conversationMiddlewareHandler{
		manager:        opts.Manager,
		cancelKeywords: opts.CancelKeywords,
		cancelReply:    opts.CancelReply,
		cancelOpts:     opts.CancelOpts,
		cancelText:     opts.CancelText,
	}
}

func ConvMiddleware(manager *conv.Manager, cancelKeywords ...string) Handler {
	return Conv(ConvOpts{
		Manager:        manager,
		CancelKeywords: cancelKeywords,
	})
}

type conversationMiddlewareHandler struct {
	manager        *conv.Manager
	cancelKeywords []string
	cancelReply    bool
	cancelOpts     *adapter.SendOpts
	cancelText     string
}

func (m *conversationMiddlewareHandler) CheckUpdate(ctx *adapter.Context, update *adapter.Update) error {
	if m.manager == nil || update.EffectiveMessage == nil {
		return ContinueGroups
	}

	msg := update.EffectiveMessage
	// Skip if it's a command
	if msg.Message != nil && msg.Text != "" && len(msg.Text) > 0 && msg.Text[0] == '/' {
		return ContinueGroups
	}

	chatID := update.ChatID()
	userID := update.UserID()
	if chatID == 0 {
		return ContinueGroups
	}

	key := conv.Key{ChatID: chatID, UserID: userID}
	state, err := m.manager.LoadState(key, update.EffectiveMessage, update)
	if err != nil || state == nil {
		return ContinueGroups
	}

	// Handle cancel keywords
	if msg.Message != nil && msg.Text != "" && len(m.cancelKeywords) > 0 {
		if slices.Contains(m.cancelKeywords, msg.Text) {
			success, wasActive, _ := update.EndConv()
			if success && wasActive && m.cancelText != "" {
				if m.cancelReply {
					_, _ = update.Reply(m.cancelText, m.cancelOpts)
				} else {
					_, _ = update.SendMessage(0, m.cancelText, m.cancelOpts)
				}
			}
			return EndGroups
		}
	}

	// Apply filter
	if filter := m.manager.GetFilter(key); filter != nil {
		if !filter(update.EffectiveMessage) {
			return ContinueGroups
		}
	}

	handler, ok := m.manager.GetStepHandler(state.Step())
	if !ok {
		return ContinueGroups
	}

	// Set up callbacks
	state.SendFn = func(text string) error {
		_, err := update.SendMessage(0, text, nil)
		return err
	}

	state.ReplyFn = func(text string) error {
		_, err := update.Reply(text, nil)
		return err
	}

	state.MediaFn = func(media tg.InputMediaClass, caption string) error {
		_, err := update.SendMedia(media, caption, nil)
		return err
	}

	// Execute handler
	if err := handler(state); err != nil {
		return err
	}

	return EndGroups
}
