package i18n

import (
	"embed"
	"fmt"

	"github.com/pageton/gotg/storage"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

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
			ext := ".ftl"
			if dir != "" && dir != "." {
				path := fmt.Sprintf("%s/%s%s", dir, lang, ext)
				content, err = fs.ReadFile(path)
			}
			if err != nil || dir == "" || dir == "." {
				path := fmt.Sprintf("%s%s", lang, ext)
				content, err = fs.ReadFile(path)
			}
		case FormatYAML:
			ext := ".yaml"
			if dir != "" && dir != "." {
				path := fmt.Sprintf("%s/%s%s", dir, lang, ext)
				content, err = fs.ReadFile(path)
			}
			if err != nil || dir == "" || dir == "." {
				path := fmt.Sprintf("%s%s", lang, ext)
				content, err = fs.ReadFile(path)
			}
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
