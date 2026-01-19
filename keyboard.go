package gotg

import (
	"github.com/gotd/td/tg"
)

// KeyboardBuilder is a fluent interface for building Telegram keyboards.
type KeyboardBuilder struct {
	rows [][]tg.KeyboardButtonClass
	row  []tg.KeyboardButtonClass
}

// Keyboard creates a new keyboard builder.
func Keyboard() *KeyboardBuilder {
	return &KeyboardBuilder{
		rows: make([][]tg.KeyboardButtonClass, 0),
		row:  make([]tg.KeyboardButtonClass, 0),
	}
}

// add adds a button to the current row.
func (k *KeyboardBuilder) add(btn tg.KeyboardButtonClass) *KeyboardBuilder {
	k.row = append(k.row, btn)
	return k
}

// Text adds a simple text button (for reply keyboards).
func (k *KeyboardBuilder) Text(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButton{Text: text})
}

// Button adds a callback button with data (for inline keyboards).
func (k *KeyboardBuilder) Button(text, data string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonCallback{
		Text: text,
		Data: []byte(data),
	})
}

// URL adds a URL button.
func (k *KeyboardBuilder) URL(text, url string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonURL{
		Text: text,
		URL:  url,
	})
}

// Switch adds a switch to inline button.
// If samePeer is true, the inline query will be sent in the current chat.
func (k *KeyboardBuilder) Switch(text string, samePeer bool, query string) *KeyboardBuilder {
	btn := &tg.KeyboardButtonSwitchInline{
		Text:     text,
		Query:    query,
		SamePeer: samePeer,
	}
	return k.add(btn)
}

// Copy adds a copy button that copies text to clipboard.
func (k *KeyboardBuilder) Copy(text, copyText string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonCopy{
		Text:     text,
		CopyText: copyText,
	})
}

// RequestPhone adds a button that requests the user's phone number.
func (k *KeyboardBuilder) RequestPhone(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonRequestPhone{Text: text})
}

// RequestGeo adds a button that requests the user's location.
func (k *KeyboardBuilder) RequestGeo(text string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonRequestGeoLocation{Text: text})
}

// RequestPoll adds a button that creates a poll.
func (k *KeyboardBuilder) RequestPoll(text string, quiz bool) *KeyboardBuilder {
	btn := &tg.KeyboardButtonRequestPoll{
		Text: text,
		Quiz: quiz,
	}
	return k.add(btn)
}

// WebApp adds a web app button.
func (k *KeyboardBuilder) WebApp(text, url string) *KeyboardBuilder {
	return k.add(&tg.KeyboardButtonWebView{
		Text: text,
		URL:  url,
	})
}

// Next moves to the next row.
func (k *KeyboardBuilder) Next() *KeyboardBuilder {
	if len(k.row) > 0 {
		k.rows = append(k.rows, k.row)
		k.row = make([]tg.KeyboardButtonClass, 0)
	}
	return k
}

// buildRows finalizes the rows for building.
func (k *KeyboardBuilder) buildRows() [][]tg.KeyboardButtonClass {
	k.Next()
	if len(k.rows) == 0 {
		return nil
	}
	return k.rows
}

// Build builds an inline keyboard (for use with inline buttons).
func (k *KeyboardBuilder) Build() tg.ReplyMarkupClass {
	rows := k.buildRows()
	if rows == nil {
		return nil
	}
	return &tg.ReplyInlineMarkup{
		Rows: makeRows(rows),
	}
}

// ReplyOptions are options for building reply keyboards.
type ReplyOptions struct {
	Resize      bool
	OneTime     bool
	Selective   bool
	Persistent  bool
	Placeholder string
}

// BuildReply builds a reply keyboard (for regular buttons).
func (k *KeyboardBuilder) BuildReply(opts ...ReplyOptions) tg.ReplyMarkupClass {
	rows := k.buildRows()
	if rows == nil {
		return nil
	}

	options := ReplyOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	markup := &tg.ReplyKeyboardMarkup{
		Rows:       makeRows(rows),
		Resize:     options.Resize,
		SingleUse:  options.OneTime,
		Selective:  options.Selective,
		Persistent: options.Persistent,
	}

	if options.Placeholder != "" {
		markup.Placeholder = options.Placeholder
	}

	return markup
}

// makeRows converts [][]tg.KeyboardButtonClass to []tg.KeyboardButtonRow.
func makeRows(rows [][]tg.KeyboardButtonClass) []tg.KeyboardButtonRow {
	result := make([]tg.KeyboardButtonRow, len(rows))
	for i, row := range rows {
		result[i] = tg.KeyboardButtonRow{Buttons: row}
	}
	return result
}
