package i18n

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"golang.org/x/text/language"
)

// GetLang retrieves the user's preferred language
func (t *Translator) GetLang(userID int64) language.Tag {
	if t.session == nil {
		return t.defaultLang
	}

	peer := t.session.GetPeerByID(userID)
	if peer != nil && peer.Language != "" {
		lang, _ := language.Parse(peer.Language)
		if _, exists := t.locales[lang]; exists {
			return lang
		}
	}

	return t.defaultLang
}

// DebugGetLang returns debug info about language retrieval
func (t *Translator) DebugGetLang(userID int64) (string, string, language.Tag) {
	if t.session == nil {
		return "no session", "", t.defaultLang
	}

	peer := t.session.GetPeerByID(userID)
	if peer == nil {
		return "peer nil", "", t.defaultLang
	}

	if peer.Language == "" {
		return "peer lang empty", "", t.defaultLang
	}

	lang, err := language.Parse(peer.Language)
	if err != nil {
		return "parse error: " + err.Error(), peer.Language, t.defaultLang
	}

	if _, exists := t.locales[lang]; exists {
		return "found", peer.Language, lang
	}

	return "locale not loaded", peer.Language, t.defaultLang
}

// HasLangSet checks if the user has an explicit language preference set
func (t *Translator) HasLangSet(userID int64) bool {
	if t.session == nil {
		return false
	}

	peer := t.session.GetPeerByID(userID)
	return peer != nil && peer.Language != ""
}

// SetLang stores the user's preferred language
func (t *Translator) SetLang(userID int64, lang language.Tag) {
	if t.session == nil {
		return
	}

	t.session.SetPeerLanguage(userID, lang.String())
}

// SetGender sets the gender context for a user
func (t *Translator) SetGender(userID int64, gender string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.genderContext[userID] = gender
}

// GetGender retrieves the gender for a user
func (t *Translator) GetGender(userID int64) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if g, ok := t.genderContext[userID]; ok {
		return g
	}
	return "other"
}

// Get retrieves a translation for the given key and user
func (t *Translator) Get(userID int64, key string, args ...any) string {
	lang := t.GetLang(userID)
	return t.GetLangCtx(lang, key, args...)
}

// Args provides context for translation variants
type Args struct {
	Count  int
	Gender string
	Args   map[string]any
}

// GetCtx retrieves a translation with context
func (t *Translator) GetCtx(userID int64, key string, ctx *Args) string {
	lang := t.GetLang(userID)
	t.mu.RLock()
	locale, exists := t.locales[lang]
	t.mu.RUnlock()

	if !exists {
		locale = t.locales[t.defaultLang]
	}
	if locale == nil {
		return key
	}

	msg, exists := locale.Messages[key]
	if !exists {
		if lang != t.defaultLang {
			t.mu.RLock()
			defaultLocale, defaultExists := t.locales[t.defaultLang]
			t.mu.RUnlock()
			if defaultExists {
				if defaultMsg, found := defaultLocale.Messages[key]; found {
					if len(defaultMsg.Variants) > 0 {
						variantKey := t.selectVariant(t.defaultLang, userID, ctx)
						if variant, ok := defaultMsg.Variants[variantKey]; ok {
							return t.formatMessage(variant, ctx)
						}
					}
					return t.formatMessage(defaultMsg.Value, ctx)
				}
			}
		}
		return key
	}

	if len(msg.Variants) > 0 {
		variantKey := t.selectVariant(lang, userID, ctx)
		if variant, ok := msg.Variants[variantKey]; ok {
			return t.formatMessage(variant, ctx)
		}
	}

	return t.formatMessage(msg.Value, ctx)
}

// GetLangCtx retrieves a translation for a specific language
func (t *Translator) GetLangCtx(lang language.Tag, key string, args ...any) string {
	t.mu.RLock()
	locale, exists := t.locales[lang]
	t.mu.RUnlock()

	if !exists {
		locale = t.locales[t.defaultLang]
	}
	if locale == nil {
		return key
	}

	msg, exists := locale.Messages[key]
	if !exists {
		if lang != t.defaultLang {
			t.mu.RLock()
			defaultLocale, defaultExists := t.locales[t.defaultLang]
			t.mu.RUnlock()
			if defaultExists {
				if defaultMsg, found := defaultLocale.Messages[key]; found {
					msg = defaultMsg
				}
			}
		}
		if msg == nil {
			return key
		}
	}

	result := msg.Value
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", arg))
	}

	return result
}

// selectVariant selects the appropriate variant based on context
func (t *Translator) selectVariant(lang language.Tag, userID int64, ctx *Args) string {
	if ctx != nil && ctx.Gender != "" {
		return ctx.Gender
	}

	if gender := t.GetGender(userID); gender != "other" {
		if ctx != nil {
			for k := range ctx.Args {
				if k == gender {
					return gender
				}
			}
		}
	}

	if ctx != nil && ctx.Count != 0 {
		return t.pluralizer.GetVariant(lang, ctx.Count)
	}

	return "other"
}

// formatMessage formats a message with context variables
func (t *Translator) formatMessage(template string, ctx *Args) string {
	result := template

	if ctx != nil {
		for key, value := range ctx.Args {
			placeholder := fmt.Sprintf("{%s}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}

		if ctx.Count != 0 {
			result = strings.ReplaceAll(result, "{count}", fmt.Sprintf("%d", ctx.Count))
		}
	}

	return result
}

// HasLang checks if a language is supported
func (t *Translator) HasLang(lang language.Tag) bool {
	t.mu.RLock()
	_, exists := t.locales[lang]
	t.mu.RUnlock()
	return exists
}

// GetSupportedLangs returns all supported languages
func (t *Translator) GetSupportedLangs() []language.Tag {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langs := make([]language.Tag, 0, len(t.locales))
	for lang := range t.locales {
		langs = append(langs, lang)
	}
	return langs
}

// User represents a user for language detection from Telegram
type User struct {
	ID           int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

// DetectLanguage detects language from Telegram user
func (t *Translator) DetectLanguage(user *User) language.Tag {
	if user.LanguageCode != "" {
		lang, _ := language.Parse(user.LanguageCode)
		if t.HasLang(lang) {
			return lang
		}
		base, _ := lang.Base()
		baseTag := language.MustParse(base.String())
		if t.HasLang(baseTag) {
			return baseTag
		}
	}
	return t.defaultLang
}

// InitFromUser initializes user language from Telegram user data
// Only sets the language if the user doesn't already have a preference saved
func (t *Translator) InitFromUser(user *tg.User) {
	if user == nil {
		return
	}

	if t.HasLangSet(user.ID) {
		return
	}

	telegramUser := &User{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LangCode,
	}

	detected := t.DetectLanguage(telegramUser)
	t.SetLang(telegramUser.ID, detected)
}

// GetWithLang retrieves a translation for a specific language (helper for manual translation)
func (t *Translator) GetWithLang(langTag string, key string, args ...any) string {
	lang, _ := language.Parse(langTag)
	return t.GetLangCtx(lang, key, args...)
}

// GetWithCtx retrieves a translation with context for a specific language
func (t *Translator) GetWithCtx(langTag string, key string, ctx *Args) string {
	lang, _ := language.Parse(langTag)
	t.mu.RLock()
	locale, exists := t.locales[lang]
	t.mu.RUnlock()

	if !exists {
		locale = t.locales[t.defaultLang]
	}
	if locale == nil {
		return key
	}

	msg, exists := locale.Messages[key]
	if !exists {
		return key
	}

	if len(msg.Variants) > 0 {
		variantKey := t.selectVariant(lang, 0, ctx)
		if variant, ok := msg.Variants[variantKey]; ok {
			return t.formatMessage(variant, ctx)
		}
	}

	return t.formatMessage(msg.Value, ctx)
}

// DebugLocaleLoadInfo returns information about what locales were loaded
func (t *Translator) DebugLocaleLoadInfo() (loadedLangs []language.Tag, loadedCount int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langs := make([]language.Tag, 0, len(t.locales))
	for lang := range t.locales {
		langs = append(langs, lang)
	}
	return langs, len(t.locales)
}
