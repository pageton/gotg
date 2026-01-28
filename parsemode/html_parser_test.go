package parsemode

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLParser_Parse_Bold(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLen  int
	}{
		{
			name:     "simple bold",
			input:    "<b>hello</b>",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "strong tag",
			input:    "<strong>hello</strong>",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "nested",
			input:    "<b><i>hello</i></b>",
			wantText: "hello",
			wantLen:  2,
		},
		{
			name:     "partial",
			input:    "hello <b>world</b>",
			wantText: "hello world",
			wantLen:  1,
		},
		{
			name:     "multiple",
			input:    "<b>hello</b> <b>world</b>",
			wantText: "hello world",
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)
			assert.Len(t, result.Entities, tt.wantLen)

			if tt.wantLen > 0 {
				bold := findEntity[*tg.MessageEntityBold](result.Entities)
				require.NotNil(t, bold)
			}
		})
	}
}

func TestHTMLParser_Parse_Italic(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLen  int
	}{
		{
			name:     "simple italic",
			input:    "<i>hello</i>",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "em tag",
			input:    "<em>hello</em>",
			wantText: "hello",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)
			assert.Len(t, result.Entities, tt.wantLen)

			if tt.wantLen > 0 {
				italic := findEntity[*tg.MessageEntityItalic](result.Entities)
				require.NotNil(t, italic)
			}
		})
	}
}

func TestHTMLParser_Parse_Underline(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "u tag",
			input:    "<u>hello</u>",
			wantText: "hello",
		},
		{
			name:     "ins tag",
			input:    "<ins>hello</ins>",
			wantText: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)

			underline := findEntity[*tg.MessageEntityUnderline](result.Entities)
			require.NotNil(t, underline)
		})
	}
}

func TestHTMLParser_Parse_Strike(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "s tag",
			input:    "<s>hello</s>",
			wantText: "hello",
		},
		{
			name:     "strike tag",
			input:    "<strike>hello</strike>",
			wantText: "hello",
		},
		{
			name:     "del tag",
			input:    "<del>hello</del>",
			wantText: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)

			strike := findEntity[*tg.MessageEntityStrike](result.Entities)
			require.NotNil(t, strike)
		})
	}
}

func TestHTMLParser_Parse_Code(t *testing.T) {
	parser := NewHTMLParser()

	result, err := parser.Parse("<code>hello</code>")
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Text)

	code := findEntity[*tg.MessageEntityCode](result.Entities)
	require.NotNil(t, code)
	assert.Equal(t, 0, code.Offset)
	assert.Equal(t, 5, code.Length)
}

func TestHTMLParser_Parse_Pre(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLang string
	}{
		{
			name:     "without language",
			input:    "<pre>hello</pre>",
			wantText: "hello",
			wantLang: "",
		},
		{
			name:     "with language",
			input:    `<pre class="language-go">hello</pre>`,
			wantText: "hello",
			wantLang: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)

			pre := findEntity[*tg.MessageEntityPre](result.Entities)
			require.NotNil(t, pre)
			assert.Equal(t, tt.wantLang, pre.Language)
		})
	}
}

func TestHTMLParser_Parse_Spoiler(t *testing.T) {
	parser := NewHTMLParser()

	result, err := parser.Parse("<spoiler>secret</spoiler>")
	require.NoError(t, err)
	assert.Equal(t, "secret", result.Text)

	spoiler := findEntity[*tg.MessageEntitySpoiler](result.Entities)
	require.NotNil(t, spoiler)
	assert.Equal(t, 0, spoiler.Offset)
	assert.Equal(t, 6, spoiler.Length)
}

func TestHTMLParser_Parse_Blockquote(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name          string
		input         string
		wantText      string
		wantCollapsed bool
	}{
		{
			name:          "quote tag",
			input:         "<quote>hello</quote>",
			wantText:      "hello",
			wantCollapsed: false,
		},
		{
			name:          "blockquote tag",
			input:         "<blockquote>hello</blockquote>",
			wantText:      "hello",
			wantCollapsed: false,
		},
		{
			name:          "collapsed",
			input:         `<blockquote collapsed="true">hello</blockquote>`,
			wantText:      "hello",
			wantCollapsed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)

			quote := findEntity[*tg.MessageEntityBlockquote](result.Entities)
			require.NotNil(t, quote)
			assert.Equal(t, tt.wantCollapsed, quote.Collapsed)
		})
	}
}

func TestHTMLParser_Parse_Link(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantType any
		wantURL  string
	}{
		{
			name:     "url link",
			input:    `<a href="https://example.com">link</a>`,
			wantText: "link",
			wantType: &tg.MessageEntityTextURL{},
			wantURL:  "https://example.com",
		},
		{
			name:     "email link",
			input:    `<a href="mailto:test@example.com">email</a>`,
			wantText: "email",
			wantType: &tg.MessageEntityEmail{},
		},
		{
			name:     "empty href",
			input:    `<a href="">url</a>`,
			wantText: "url",
			wantType: &tg.MessageEntityURL{},
		},
		{
			name:     "protocol-less t.me link",
			input:    `<a href="t.me/l9l9l">Team David</a>`,
			wantText: "Team David",
			wantType: &tg.MessageEntityTextURL{},
			wantURL:  "https://t.me/l9l9l",
		},
		{
			name:     "protocol-less domain link",
			input:    `<a href="example.com/path">link</a>`,
			wantText: "link",
			wantType: &tg.MessageEntityTextURL{},
			wantURL:  "https://example.com/path",
		},
		{
			name:     "protocol-relative link",
			input:    `<a href="//example.com/path">link</a>`,
			wantText: "link",
			wantType: &tg.MessageEntityTextURL{},
			wantURL:  "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)
			assert.Len(t, result.Entities, 1)

			entity := result.Entities[0]
			assert.IsType(t, tt.wantType, entity)

			if tt.wantURL != "" {
				if textURL, ok := entity.(*tg.MessageEntityTextURL); ok {
					assert.Equal(t, tt.wantURL, textURL.URL)
				}
			}
		})
	}
}

func TestHTMLParser_Parse_CustomEmoji(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name        string
		input       string
		wantText    string
		wantEmojiID int64
	}{
		{
			name:        "emoji tag",
			input:       `<emoji id="12345">emoji</emoji>`,
			wantText:    "emoji",
			wantEmojiID: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)

			emoji := findEntity[*tg.MessageEntityCustomEmoji](result.Entities)
			require.NotNil(t, emoji)
			assert.Equal(t, tt.wantEmojiID, emoji.DocumentID)
		})
	}
}

func TestHTMLParser_Parse_UnsupportedTags(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "div tag",
			input:    "<div>hello</div>",
			wantText: "<div>hello</div>",
		},
		{
			name:     "span tag",
			input:    "<span>hello</span>",
			wantText: "<span>hello</span>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)
			assert.Nil(t, result.Entities)
		})
	}
}

func TestHTMLParser_Parse_HTMLUnescape(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "less than",
			input:    "&lt;",
			wantText: "<",
		},
		{
			name:     "greater than",
			input:    "&gt;",
			wantText: ">",
		},
		{
			name:     "ampersand",
			input:    "&amp;",
			wantText: "&",
		},
		{
			name:     "quoted",
			input:    "&quot;",
			wantText: "\"",
		},
		{
			name:     "mixed with tags",
			input:    "<b>&lt;hello&gt;</b>",
			wantText: "<hello>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, result.Text)
		})
	}
}

func TestHTMLParser_Parse_Empty(t *testing.T) {
	parser := NewHTMLParser()

	result, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", result.Text)
	assert.Nil(t, result.Entities)
}

func TestHTMLParser_Parse_MultipleEntities(t *testing.T) {
	parser := NewHTMLParser()

	input := `<b>bold</b> <i>italic</i> <u>underline</u> <s>strike</s> <code>code</code>`
	result, err := parser.Parse(input)
	require.NoError(t, err)
	assert.Equal(t, "bold italic underline strike code", result.Text)
	assert.Len(t, result.Entities, 5)
}

// Helper function to find an entity of a specific type
func findEntity[T any](entities []tg.MessageEntityClass) T {
	var zero T
	for _, e := range entities {
		if typed, ok := e.(T); ok {
			return typed
		}
	}
	return zero
}
