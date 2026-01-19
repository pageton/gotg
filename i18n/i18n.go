package i18n

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// LocaleFormat represents the format of locale files
type LocaleFormat string

const (
	FormatFTL  LocaleFormat = "ftl"
	FormatYAML LocaleFormat = "yaml"
)

// Translator handles internationalization with support for multiple formats
type Translator struct {
	mu            sync.RWMutex
	locales       map[language.Tag]*Locale
	defaultLang   language.Tag
	session       *storage.PeerStorage
	format        LocaleFormat
	pluralizer    *Pluralizer
	genderContext map[int64]string // userID -> gender
}

// Locale contains translations for a specific language
type Locale struct {
	Lang     language.Tag
	Messages map[string]*Message
	FTL      map[string]string // Raw FTL messages if format is FTL
}

// Message represents a translatable message with variants
type Message struct {
	Key      string
	Value    string
	Variants map[string]string // For pluralization, gender, etc.
	Attrs    map[string]string
}

// LocaleConfig configures the translator
type LocaleConfig struct {
	DefaultLang    language.Tag
	Session        *storage.PeerStorage
	Format         LocaleFormat
	EmbedFS        embed.FS
	LocaleDir      string
	SupportedLangs []language.Tag
}

// NewTranslator creates a new i18n translator
func NewTranslator(config *LocaleConfig) *Translator {
	t := &Translator{
		locales:       make(map[language.Tag]*Locale),
		defaultLang:   config.DefaultLang,
		session:       config.Session,
		format:        config.Format,
		pluralizer:    NewPluralizer(),
		genderContext: make(map[int64]string),
	}

	if len(config.SupportedLangs) > 0 {
		t.loadEmbeddedLocales(config.EmbedFS, config.LocaleDir, config.SupportedLangs)
	}

	return t
}

// SetSession sets or updates the peer storage for the translator
// This can be used to set the session after client creation
func (t *Translator) SetSession(session *storage.PeerStorage) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.session = session
}

// loadEmbeddedLocales loads locale files from embedded FS
func (t *Translator) loadEmbeddedLocales(fs embed.FS, dir string, langs []language.Tag) {
	for _, lang := range langs {
		var content []byte
		var err error

		switch t.format {
		case FormatFTL:
			path := fmt.Sprintf("%s/%s.ftl", dir, lang)
			content, err = fs.ReadFile(path)
		case FormatYAML:
			path := fmt.Sprintf("%s/%s.yaml", dir, lang)
			content, err = fs.ReadFile(path)
		}

		if err != nil {
			continue
		}

		if t.format == FormatFTL {
			t.LoadFTL(lang, string(content))
		} else {
			t.LoadYAML(lang, content)
		}
	}
}

// LoadYAML loads YAML locale content
func (t *Translator) LoadYAML(lang language.Tag, content []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var data map[string]any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	locale := &Locale{
		Lang:     lang,
		Messages: make(map[string]*Message),
	}

	t.parseYAMLMap(data, "", locale.Messages)
	t.locales[lang] = locale
	return nil
}

// parseYAMLMap recursively parses YAML map into messages
func (t *Translator) parseYAMLMap(data map[string]any, prefix string, messages map[string]*Message) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			t.parseYAMLMap(v, fullKey, messages)
		case string:
			messages[fullKey] = &Message{
				Key:   fullKey,
				Value: v,
			}
		case map[string]string:
			// Handle variants (pluralization, gender, etc.)
			msg := &Message{
				Key:      fullKey,
				Variants: make(map[string]string),
			}
			for variantKey, variantValue := range v {
				if variantKey == "other" || variantKey == "base" {
					msg.Value = variantValue
				} else {
					msg.Variants[variantKey] = variantValue
				}
			}
			if msg.Value == "" && len(v) > 0 {
				// Use first variant as base
				for _, val := range v {
					msg.Value = val
					break
				}
			}
			messages[fullKey] = msg
		}
	}
}

// LoadFTL loads FTL locale content
func (t *Translator) LoadFTL(lang language.Tag, content string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	locale := &Locale{
		Lang:     lang,
		Messages: make(map[string]*Message),
		FTL:      make(map[string]string),
	}

	parser := NewFTLParser()
	messages, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse FTL: %w", err)
	}

	for key, msg := range messages {
		locale.Messages[key] = msg
		locale.FTL[key] = msg.Value
	}

	t.locales[lang] = locale
	return nil
}

// GetLang retrieves the user's preferred language
func (t *Translator) GetLang(userID int64) language.Tag {
	if t.session == nil {
		return t.defaultLang
	}

	// Check session for stored language
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

// GetCtx retrieves a translation with context
type Args struct {
	Count  int
	Gender string
	Args   map[string]any
}

func (t *Translator) GetCtx(userID int64, key string, ctx *Args) string {
	lang := t.GetLang(userID)
	t.mu.RLock()
	locale, exists := t.locales[lang]
	t.mu.RUnlock()

	if !exists {
		// Fallback to default language if locale doesn't exist
		locale = t.locales[t.defaultLang]
	}
	if locale == nil {
		return key
	}

	msg, exists := locale.Messages[key]
	if !exists {
		// Fallback to default language if key doesn't exist in user's language
		if lang != t.defaultLang {
			t.mu.RLock()
			defaultLocale, defaultExists := t.locales[t.defaultLang]
			t.mu.RUnlock()
			if defaultExists {
				if defaultMsg, found := defaultLocale.Messages[key]; found {
					// Check for variants in default locale
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

	// Check for variants based on context
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
		// Fallback to default language if locale doesn't exist
		locale = t.locales[t.defaultLang]
	}
	if locale == nil {
		return key
	}

	msg, exists := locale.Messages[key]
	if !exists {
		// Fallback to default language if key doesn't exist in user's language
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
	// Gender-based selection
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

	// Pluralization-based selection
	if ctx != nil && ctx.Count != 0 {
		return t.pluralizer.GetVariant(lang, ctx.Count)
	}

	return "other"
}

// formatMessage formats a message with context variables
func (t *Translator) formatMessage(template string, ctx *Args) string {
	result := template

	if ctx != nil {
		// Replace named placeholders
		for key, value := range ctx.Args {
			placeholder := fmt.Sprintf("{%s}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}

		// Replace {count} if count is set
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
		// Try base language (e.g., 'en' from 'en-US')
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

	// Check if user already has a language preference set
	if t.HasLangSet(user.ID) {
		// User already has a language set, don't override it
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

	// Check for variants based on context
	if len(msg.Variants) > 0 {
		variantKey := t.selectVariant(lang, 0, ctx)
		if variant, ok := msg.Variants[variantKey]; ok {
			return t.formatMessage(variant, ctx)
		}
	}

	return t.formatMessage(msg.Value, ctx)
}
