package parsemode

import "strings"

var mdV2Escapes [128]string

func init() {
	for _, r := range []rune{'_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'} {
		mdV2Escapes[r] = `\` + string(r)
	}
}

func EscapeMarkdownV2(text string) string {
	var b strings.Builder
	b.Grow(len(text) * 2)
	for _, r := range text {
		if r < 128 && mdV2Escapes[r] != "" {
			b.WriteString(mdV2Escapes[r])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func EscapeHTML(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
