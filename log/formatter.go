package log

import (
	"fmt"
	"strconv"
	"strings"
)

type Formatter interface {
	Format(rec *Record) []byte
}

type TextFormatter struct {
	Color      bool
	Timestamp  bool
	TimeLayout string
	Caller     bool
	FuncName   bool
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		Color:      true,
		Timestamp:  true,
		TimeLayout: "15:04:05",
		Caller:     true,
	}
}

func (f *TextFormatter) Format(rec *Record) []byte {
	// Pre-allocate enough for typical log line
	var b strings.Builder
	b.Grow(192)

	lvl := rec.Level
	icon := lvl.Icon()
	tag := lvl.Tag()

	// === Timestamp (dimmed) ===
	if f.Timestamp {
		if f.Color {
			b.WriteString(cDim)
		}
		b.WriteString(rec.Time.Format(f.TimeLayout))
		if f.Color {
			b.WriteString(cReset)
		}
		b.WriteByte(' ')
	}

	// === Icon + Tag (colored) ===
	if f.Color {
		b.WriteString(lvl.Color())
	}
	b.WriteString(icon)
	b.WriteByte(' ')
	b.WriteString(tag)
	// Pad tag to 5 chars for alignment (FATAL=5, DEBUG=5, ERROR=5, WARN=4+1, INFO=4+1, OK=2+3)
	for i := len(tag); i < 5; i++ {
		b.WriteByte(' ')
	}
	if f.Color {
		b.WriteString(cReset)
	}

	// === Message (bright) ===
	b.WriteString("  ")
	if f.Color && (lvl == LevelError || lvl == LevelFatal) {
		b.WriteString(cBold)
	}
	b.WriteString(rec.Message)
	if f.Color && (lvl == LevelError || lvl == LevelFatal) {
		b.WriteString(cReset)
	}

	// === Module ===
	if rec.Module != "" {
		b.WriteString("  ")
		if f.Color {
			b.WriteString(cDim)
		}
		b.WriteByte('[')
		b.WriteString(rec.Module)
		b.WriteByte(']')
		if f.Color {
			b.WriteString(cReset)
		}
	}

	// === Fields (dimmed key=value) ===
	for _, field := range rec.Fields {
		b.WriteString("  ")
		if f.Color {
			b.WriteString(cDim)
		}
		b.WriteString(field.Key)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(field.Value))
		if f.Color {
			b.WriteString(cReset)
		}
	}

	// === Caller (dimmed, appended for errors or when enabled) ===
	if f.Caller && rec.File != "" {
		b.WriteString("  ")
		if f.Color {
			b.WriteString(cDim)
		}
		b.WriteString(rec.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(rec.Line))
		if f.FuncName && rec.Func != "" {
			b.WriteByte(' ')
			b.WriteString(rec.Func)
		}
		if f.Color {
			b.WriteString(cReset)
		}
	}

	b.WriteByte('\n')
	return []byte(b.String())
}
