package parsemode

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownParser_Parse_Bold(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLen  int
	}{
		{
			name:     "double asterisk",
			input:    "**hello**",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "double underscore",
			input:    "__hello__",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "mixed",
			input:    "**bold** __also bold__",
			wantText: "bold also bold",
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

func TestMarkdownParser_Parse_Italic(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLen  int
	}{
		{
			name:     "single asterisk",
			input:    "*hello*",
			wantText: "hello",
			wantLen:  1,
		},
		{
			name:     "single underscore",
			input:    "_hello_",
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

func TestMarkdownParser_Parse_Strike(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "single tilde",
			input:    "~hello~",
			wantText: "hello",
		},
		{
			name:     "double tilde",
			input:    "~~hello~~",
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

func TestMarkdownParser_Parse_Spoiler(t *testing.T) {
	parser := NewMarkdownParser()

	result, err := parser.Parse("||secret||")
	require.NoError(t, err)
	assert.Equal(t, "secret", result.Text)

	spoiler := findEntity[*tg.MessageEntitySpoiler](result.Entities)
	require.NotNil(t, spoiler)
}

func TestMarkdownParser_Parse_Code(t *testing.T) {
	parser := NewMarkdownParser()

	result, err := parser.Parse("`hello`")
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Text)

	code := findEntity[*tg.MessageEntityCode](result.Entities)
	require.NotNil(t, code)
}

func TestMarkdownParser_Parse_CodeBlock(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantText string
		wantLang string
	}{
		{
			name:     "without language",
			input:    "```\ncode\n```",
			wantText: "code",
			wantLang: "",
		},
		{
			name:     "with language",
			input:    "```go\nfunc main() {}\n```",
			wantText: "func main() {}",
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

func TestMarkdownParser_Parse_Link(t *testing.T) {
	parser := NewMarkdownParser()

	result, err := parser.Parse("[link text](https://example.com)")
	require.NoError(t, err)
	assert.Equal(t, "link text", result.Text)

	textURL := findEntity[*tg.MessageEntityTextURL](result.Entities)
	require.NotNil(t, textURL)
	assert.Equal(t, "https://example.com", textURL.URL)
}

func TestMarkdownParser_Parse_CustomEmoji(t *testing.T) {
	parser := NewMarkdownParser()

	result, err := parser.Parse("::12345::")
	require.NoError(t, err)
	assert.Equal(t, "", result.Text) // Empty text for emoji

	emoji := findEntity[*tg.MessageEntityCustomEmoji](result.Entities)
	require.NotNil(t, emoji)
	assert.Equal(t, int64(12345), emoji.DocumentID)
}

func TestMarkdownParser_Parse_Blockquote(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name          string
		input         string
		wantText      string
		wantCollapsed bool
	}{
		{
			name:          "single line",
			input:         "> hello",
			wantText:      "hello",
			wantCollapsed: false,
		},
		{
			name:          "collapsed",
			input:         ">>> hello",
			wantText:      "hello",
			wantCollapsed: true,
		},
		{
			name:          "multiple lines",
			input:         "> hello\n> world",
			wantText:      "hello\nworld",
			wantCollapsed: false,
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

func TestMarkdownParser_Parse_EscapedCharacters(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantText string
	}{
		{
			name:     "escaped asterisk",
			input:    `\*hello\*`,
			wantText: "*hello*",
		},
		{
			name:     "escaped underscore",
			input:    `\_hello\_`,
			wantText: "_hello_",
		},
		{
			name:     "escaped backtick",
			input:    "\\`hello\\`",
			wantText: "`hello`",
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

func TestMarkdownParser_Parse_Empty(t *testing.T) {
	parser := NewMarkdownParser()

	result, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", result.Text)
	assert.Nil(t, result.Entities)
}

func TestMarkdownParser_Parse_MultipleEntities(t *testing.T) {
	parser := NewMarkdownParser()

	input := "**bold** *italic* ~strike~ ||spoiler|| `code`"
	result, err := parser.Parse(input)
	require.NoError(t, err)
	assert.Equal(t, "bold italic strike spoiler code", result.Text)
	assert.Len(t, result.Entities, 5)
}

func TestMarkdownParser_MarkdownToHTML(t *testing.T) {
	parser := NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantHTML string
	}{
		{
			name:     "bold",
			input:    "**hello**",
			wantHTML: "<b>hello</b>",
		},
		{
			name:     "italic",
			input:    "*hello*",
			wantHTML: "<i>hello</i>",
		},
		{
			name:     "code",
			input:    "`hello`",
			wantHTML: "<code>hello</code>",
		},
		{
			name:     "link",
			input:    "[text](url)",
			wantHTML: `<a href="url">text</a>`,
		},
		{
			name:     "spoiler",
			input:    "||hello||",
			wantHTML: "<spoiler>hello</spoiler>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.MarkdownToHTML(tt.input)
			assert.Equal(t, tt.wantHTML, result)
		})
	}
}

func TestMarkdownParser_Format(t *testing.T) {
	parser := NewMarkdownParser()

	// Format should escape special characters
	input := "*hello* [world]"
	result := parser.Format(input)
	// Should contain escaped characters
	assert.Contains(t, result, "\\*")
	assert.Contains(t, result, "\\[")
	assert.Contains(t, result, "\\]")
}
