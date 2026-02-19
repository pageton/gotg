package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/pageton/gotg/storage"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// SetSession sets or updates the peer storage for the translator.
// This can be used to set the session after client creation.
func (t *Translator) SetSession(session *storage.PeerStorage) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.session = session
}

// loadEmbeddedLocales loads locale files from an embedded FS.
//
// It supports two directory layouts (checked in order):
//
//  1. Directory-based: locales/en/*.yaml, locales/ar/*.yaml
//     Each language is a sub-directory whose name is the BCP-47 tag.
//     All matching files inside the directory are merged into one Locale.
//
//  2. Flat-file: locales/en.yaml, locales/ar.yaml
//     Each language is a single file named {tag}.{ext}.
//
// When langs is empty the loader auto-discovers languages by scanning the
// filesystem for directories or files whose names parse as valid BCP-47 tags.
func (t *Translator) loadEmbeddedLocales(efs embed.FS, dir string, langs []language.Tag) {
	if dir == "" || dir == "." {
		dir = "."
	}

	ext := t.fileExtension()

	// Auto-discover languages when none are explicitly listed.
	if len(langs) == 0 {
		langs = t.discoverLangs(efs, dir, ext)
	}

	for _, lang := range langs {
		// Strategy 1 — directory-based: {dir}/{lang}/*.{ext}
		langDir := path.Join(dir, lang.String())
		if entries, err := fs.ReadDir(efs, langDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), "."+ext) {
					continue
				}
				filePath := path.Join(langDir, entry.Name())
				content, err := efs.ReadFile(filePath)
				if err != nil {
					continue
				}
				t.loadContent(lang, content)
			}
			continue // skip flat-file fallback if the directory existed
		}

		// Strategy 2 — flat file: {dir}/{lang}.{ext}
		filePath := path.Join(dir, lang.String()+"."+ext)
		content, err := efs.ReadFile(filePath)
		if err != nil {
			// Also try without directory prefix for root embeds.
			content, err = efs.ReadFile(lang.String() + "." + ext)
			if err != nil {
				continue
			}
		}
		t.loadContent(lang, content)
	}
}

// discoverLangs scans the embedded FS for directories or files that represent
// valid BCP-47 language tags.
func (t *Translator) discoverLangs(efs embed.FS, dir, ext string) []language.Tag {
	entries, err := fs.ReadDir(efs, dir)
	if err != nil {
		return nil
	}

	var langs []language.Tag
	seen := make(map[language.Tag]bool)

	for _, entry := range entries {
		name := entry.Name()

		var tagStr string
		if entry.IsDir() {
			tagStr = name
		} else if strings.HasSuffix(name, "."+ext) {
			tagStr = strings.TrimSuffix(name, "."+ext)
		} else {
			continue
		}

		tag, err := language.Parse(tagStr)
		if err != nil {
			continue
		}
		if !seen[tag] {
			seen[tag] = true
			langs = append(langs, tag)
		}
	}
	return langs
}

// loadContent parses content according to the translator's format and merges
// it into the locale for the given language tag. Calling this multiple times
// for the same language merges (not replaces) the messages.
func (t *Translator) loadContent(lang language.Tag, content []byte) {
	switch t.format {
	case FormatFTL:
		t.loadFTLContent(lang, string(content))
	default:
		t.loadYAMLContent(lang, content)
	}
}

// fileExtension returns the file extension for the current format.
func (t *Translator) fileExtension() string {
	switch t.format {
	case FormatFTL:
		return "ftl"
	default:
		return "yaml"
	}
}

// ---------------------------------------------------------------------------
// YAML
// ---------------------------------------------------------------------------

// LoadYAML loads YAML locale content for a language.
// If the language already has messages loaded, new keys are merged in
// (existing keys are overwritten).
func (t *Translator) LoadYAML(lang language.Tag, content []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.loadYAMLLocked(lang, content)
}

func (t *Translator) loadYAMLContent(lang language.Tag, content []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_ = t.loadYAMLLocked(lang, content)
}

func (t *Translator) loadYAMLLocked(lang language.Tag, content []byte) error {
	var data map[string]any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	locale := t.getOrCreateLocale(lang)
	t.parseYAMLMap(data, "", locale.Messages)
	return nil
}

// parseYAMLMap recursively parses a YAML map into messages.
func (t *Translator) parseYAMLMap(data map[string]any, prefix string, messages map[string]*Message) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			// Check if this is a plural/variant map (all string values with
			// known plural keys like "one", "other", "few", "many", etc.)
			if isVariantMap(v) {
				msg := &Message{
					Key:      fullKey,
					Variants: make(map[string]string),
				}
				for variantKey, variantValue := range v {
					sv := fmt.Sprintf("%v", variantValue)
					if variantKey == "other" || variantKey == "base" {
						msg.Value = sv
					} else {
						msg.Variants[variantKey] = sv
					}
				}
				if msg.Value == "" && len(v) > 0 {
					for _, val := range v {
						msg.Value = fmt.Sprintf("%v", val)
						break
					}
				}
				messages[fullKey] = msg
			} else {
				t.parseYAMLMap(v, fullKey, messages)
			}
		case string:
			messages[fullKey] = &Message{
				Key:   fullKey,
				Value: v,
			}
		case map[string]string:
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
				for _, val := range v {
					msg.Value = val
					break
				}
			}
			messages[fullKey] = msg
		}
	}
}

// knownPluralForms are the CLDR plural category names.
var knownPluralForms = map[string]bool{
	"zero": true, "one": true, "two": true,
	"few": true, "many": true, "other": true, "base": true,
}

// isVariantMap returns true when every key in the map is a known plural form
// and every value is a string. This distinguishes:
//
//	items_count: { one: "...", other: "..." }   → variant map
//	features:    { formatting: "...", i18n: { ... } } → nested map
func isVariantMap(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}
	for k, v := range m {
		if !knownPluralForms[k] {
			return false
		}
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// FTL
// ---------------------------------------------------------------------------

// LoadFTL loads FTL locale content for a language.
// If the language already has messages loaded, new keys are merged in
// (existing keys are overwritten).
func (t *Translator) LoadFTL(lang language.Tag, content string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.loadFTLLocked(lang, content)
}

func (t *Translator) loadFTLContent(lang language.Tag, content string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_ = t.loadFTLLocked(lang, content)
}

func (t *Translator) loadFTLLocked(lang language.Tag, content string) error {
	parser := NewFTLParser()
	messages, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse FTL: %w", err)
	}

	locale := t.getOrCreateLocale(lang)
	if locale.FTL == nil {
		locale.FTL = make(map[string]string)
	}

	for key, msg := range messages {
		locale.Messages[key] = msg
		locale.FTL[key] = msg.Value
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// getOrCreateLocale returns the existing locale for a tag or creates a new one.
// Must be called with t.mu held.
func (t *Translator) getOrCreateLocale(lang language.Tag) *Locale {
	locale, ok := t.locales[lang]
	if !ok {
		locale = &Locale{
			Lang:     lang,
			Messages: make(map[string]*Message),
		}
		t.locales[lang] = locale
	}
	return locale
}
