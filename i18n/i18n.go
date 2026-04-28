// Package i18n provides internationalization support for gotg.
// Supports Fluent (.ftl) and YAML formats with 142 CLDR plural rules.

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

// LocaleConfig configures the translator.
type LocaleConfig struct {
	DefaultLang language.Tag
	Session     *storage.PeerStorage
	Format      LocaleFormat
	EmbedFS     embed.FS
	// LocaleDir is the root directory inside EmbedFS that contains locale
	// files or sub-directories. Both layouts are supported:
	//
	//   Directory-based:  locales/en/*.yaml   locales/ar/*.yaml
	//   Flat-file:        locales/en.yaml     locales/ar.yaml
	//
	// The loader checks for a sub-directory first; if none exists it falls
	// back to the flat-file layout.
	LocaleDir string
	// SupportedLangs explicitly lists the languages to load.
	// When empty, languages are auto-discovered from LocaleDir by scanning
	// for sub-directories and files whose names are valid BCP-47 tags.
	SupportedLangs []language.Tag
}

// NewTranslator creates a new i18n translator.
func NewTranslator(config *LocaleConfig) *Translator {
	t := &Translator{
		locales:       make(map[language.Tag]*Locale),
		defaultLang:   config.DefaultLang,
		session:       config.Session,
		format:        config.Format,
		pluralizer:    NewPluralizer(),
		genderContext: make(map[int64]string),
	}

	t.loadEmbeddedLocales(config.EmbedFS, config.LocaleDir, config.SupportedLangs)

	return t
}
