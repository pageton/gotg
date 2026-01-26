package i18n

import (
	"embed"
	"sync"

	"github.com/pageton/gotg/storage"
	"golang.org/x/text/language"
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
