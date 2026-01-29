package i18n

import (
	"math"

	"golang.org/x/text/language"
)

// Pluralizer handles pluralization rules for different languages
type Pluralizer struct {
	rules map[language.Tag]PluralRule
}

// PluralRule defines the pluralization rule for a language
type PluralRule func(n int) string

// NewPluralizer creates a new pluralizer with built-in rules
func NewPluralizer() *Pluralizer {
	p := &Pluralizer{
		rules: make(map[language.Tag]PluralRule),
	}

	// Add built-in pluralization rules
	p.addBuiltInRules()

	return p
}

// addBuiltInRules adds pluralization rules for common languages
func (p *Pluralizer) addBuiltInRules() {
	// English: one, other
	p.rules[language.English] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Spanish: one, other
	p.rules[language.Spanish] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// French: one, other
	p.rules[language.French] = func(n int) string {
		if n == 0 || n == 1 {
			return "one"
		}
		return "other"
	}

	// German: one, other
	p.rules[language.German] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Italian: one, other
	p.rules[language.Italian] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Portuguese: one, other
	p.rules[language.Portuguese] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Russian: one, few, many, other
	p.rules[language.Russian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if mod10 == 1 && mod100 != 11 {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
			return "few"
		}
		if mod10 == 0 || (mod10 >= 5 && mod10 <= 9) || (mod100 >= 11 && mod100 <= 14) {
			return "many"
		}
		return "other"
	}

	// Polish: one, few, many
	p.rules[language.Polish] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		if nAbs == 1 {
			return "one"
		}
		mod10 := nAbs % 10
		mod100 := nAbs % 100
		if (mod10 >= 2 && mod10 <= 4) && !(mod100 >= 12 && mod100 <= 14) {
			return "few"
		}
		if (mod10 != 1 && nAbs != 0) && (mod10 <= 1 || mod10 >= 5) {
			return "many"
		}
		return "other"
	}

	// Ukrainian: one, few, many, other
	p.rules[language.Ukrainian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if mod10 == 1 && mod100 != 11 {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
			return "few"
		}
		if mod10 == 0 || (mod10 >= 5 && mod10 <= 9) || (mod100 >= 11 && mod100 <= 14) {
			return "many"
		}
		return "other"
	}

	// Arabic: zero, one, two, few, many, other
	p.rules[language.Arabic] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod100 := nAbs % 100

		if nAbs == 0 {
			return "zero"
		}
		if nAbs == 1 {
			return "one"
		}
		if nAbs == 2 {
			return "two"
		}
		if mod100 >= 3 && mod100 <= 10 {
			return "few"
		}
		if mod100 >= 11 && mod100 <= 99 {
			return "many"
		}
		return "other"
	}

	// Czech: one, few, many, other
	p.rules[language.Czech] = func(n int) string {
		if n == 1 {
			return "one"
		}
		if n >= 2 && n <= 4 {
			return "few"
		}
		return "other"
	}

	// Turkish: one, other
	p.rules[language.Turkish] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Japanese: other (no pluralization)
	p.rules[language.Japanese] = func(n int) string {
		return "other"
	}

	// Chinese: other (no pluralization)
	p.rules[language.Chinese] = func(n int) string {
		return "other"
	}

	// Korean: other (no pluralization)
	p.rules[language.Korean] = func(n int) string {
		return "other"
	}

	// Vietnamese: other (no pluralization)
	p.rules[language.Vietnamese] = func(n int) string {
		return "other"
	}

	// Thai: other (no pluralization)
	p.rules[language.Thai] = func(n int) string {
		return "other"
	}

	// Indonesian: other (no pluralization)
	p.rules[language.Indonesian] = func(n int) string {
		return "other"
	}

	// Hindi: one, other
	p.rules[language.Hindi] = func(n int) string {
		if n == 0 || n == 1 {
			return "one"
		}
		return "other"
	}

	// Hebrew: one, two, many, other
	p.rules[language.Hebrew] = func(n int) string {
		if n == 1 {
			return "one"
		}
		if n == 2 {
			return "two"
		}
		if n != 0 && n%10 == 0 {
			return "many"
		}
		return "other"
	}

	// Dutch: one, other
	p.rules[language.Dutch] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Swedish: one, other
	p.rules[language.Swedish] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Danish: one, other
	p.rules[language.Danish] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Norwegian: one, other
	p.rules[language.Norwegian] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Finnish: one, other
	p.rules[language.Finnish] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Greek: one, other
	p.rules[language.Greek] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Hungarian: one, other
	p.rules[language.Hungarian] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Romanian: one, few, other
	p.rules[language.Romanian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		if nAbs == 1 {
			return "one"
		}
		if nAbs == 0 || (nAbs != 1 && nAbs%100 >= 1 && nAbs%100 <= 19) {
			return "few"
		}
		return "other"
	}

	// Slovak: one, few, many, other
	p.rules[language.Slovak] = func(n int) string {
		if n == 1 {
			return "one"
		}
		if n >= 2 && n <= 4 {
			return "few"
		}
		return "other"
	}

	// Bulgarian: one, other
	p.rules[language.Bulgarian] = func(n int) string {
		if n == 1 {
			return "one"
		}
		return "other"
	}

	// Serbian: one, few, other
	p.rules[language.Serbian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if mod10 == 1 && mod100 != 11 {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
			return "few"
		}
		return "other"
	}

	// Croatian: one, few, other
	p.rules[language.Croatian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if mod10 == 1 && mod100 != 11 {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
			return "few"
		}
		return "other"
	}

	// Lithuanian: one, few, many, other
	p.rules[language.Lithuanian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if mod10 == 1 && !(mod100 >= 11 && mod100 <= 19) {
			return "one"
		}
		if mod10 >= 2 && mod10 <= 9 && !(mod100 >= 11 && mod100 <= 19) {
			return "few"
		}
		return "other"
	}

	// Latvian: zero, one, other
	p.rules[language.Latvian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod10 := nAbs % 10
		mod100 := nAbs % 100

		if nAbs == 0 || (mod10 == 0 && (mod100 < 11 || mod100 > 19)) {
			return "zero"
		}
		if nAbs == 1 || (mod10 == 1 && (mod100 < 11 || mod100 > 19)) {
			return "one"
		}
		return "other"
	}

	// Slovenian: one, two, few, other
	p.rules[language.Slovenian] = func(n int) string {
		nAbs := int(math.Abs(float64(n)))
		mod100 := nAbs % 100

		if mod100 == 1 {
			return "one"
		}
		if mod100 == 2 {
			return "two"
		}
		if mod100 >= 3 && mod100 <= 4 {
			return "few"
		}
		return "other"
	}

	// Macedonian: one, other
	p.rules[language.Macedonian] = func(n int) string {
		if n == 1 || (n%10 == 1 && n%100 != 11) {
			return "one"
		}
		return "other"
	}
}

// GetVariant returns the plural variant for a given count and language
func (p *Pluralizer) GetVariant(lang language.Tag, count int) string {
	// Try exact match
	if rule, ok := p.rules[lang]; ok {
		return rule(count)
	}

	// Try base language (e.g., 'en' from 'en-US')
	base, _ := lang.Base()
	baseTag := language.Make(base.String())
	if rule, ok := p.rules[baseTag]; ok {
		return rule(count)
	}

	// Default to English-style pluralization
	if count == 1 {
		return "one"
	}
	return "other"
}

// AddRule adds a custom pluralization rule for a language
func (p *Pluralizer) AddRule(lang language.Tag, rule PluralRule) {
	p.rules[lang] = rule
}
