package parsemode

import (
	"testing"

	"github.com/gotd/td/tg"
)

var benchStr100 = "Hello, this is a test string with special chars: _*[]()~`>#+-=|{}.! and more normal text here."

func BenchmarkEscapeMarkdownV2(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EscapeMarkdownV2(benchStr100)
	}
}

func BenchmarkEscapeHTML(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EscapeHTML(benchStr100)
	}
}

func BenchmarkCombineFormattedText(b *testing.B) {
	segments := make([]FormattedText, 10)
	for i := range segments {
		segments[i] = FormattedText{
			Text: "segment text",
			Entities: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 0, Length: 7},
			},
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Combine(segments...)
	}
}
