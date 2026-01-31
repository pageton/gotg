package log

import (
	"fmt"
	"os"
	"time"
)

// Logger is the core logging type. It is safe for concurrent use
// because thread-safety is handled by the underlying Writer.
type Logger struct {
	cfg     Config
	fmt     Formatter
	writer  Writer
	fileFmt Formatter
	fileW   Writer
}

// New creates a Logger from the given Config.
// It wires a TextFormatter and ConsoleWriter whose settings
// mirror the Config flags.
func New(cfg Config) *Logger {
	if cfg.Debug {
		cfg.MinLevel = LevelDebug
	}
	tf := &TextFormatter{
		Color:      cfg.Color,
		Timestamp:  cfg.Timestamp,
		TimeLayout: "15:04:05",
		Caller:     cfg.Caller,
		FuncName:   cfg.FuncName,
	}
	l := &Logger{
		cfg:    cfg,
		fmt:    tf,
		writer: NewConsoleWriter(),
	}
	if cfg.LogFile != "" {
		fw, err := NewFileWriter(cfg.LogFile)
		if err == nil {
			l.fileW = fw
			l.fileFmt = &TextFormatter{
				Color:      false,
				Timestamp:  true,
				TimeLayout: "2006-01-02 15:04:05",
				Caller:     cfg.Caller,
				FuncName:   cfg.FuncName,
			}
		}
	}
	return l
}

// NewWithWriter is like New but lets the caller supply a custom Writer
// (e.g. FileWriter, MultiWriter).
func NewWithWriter(cfg Config, w Writer) *Logger {
	l := New(cfg)
	l.writer = w
	return l
}

// Default returns a Logger with DefaultConfig().
func Default() *Logger {
	return New(DefaultConfig())
}

// Nop returns a Logger that discards all output.
// Used when no LogConfig is provided.
func Nop() *Logger {
	return New(Config{MinLevel: LevelOff})
}

// WithModule returns a child Logger that tags every record with the
// given module name. The child shares the same Writer and Formatter.
func (l *Logger) WithModule(name string) *Logger {
	child := *l // shallow copy
	child.cfg.Module = name
	return &child
}

// --- structured log methods (key-value pairs) ---

func (l *Logger) Debug(msg string, kvs ...any) {
	l.emit(LevelDebug, 3, msg, kvs)
}

func (l *Logger) Info(msg string, kvs ...any) {
	l.emit(LevelInfo, 3, msg, kvs)
}

func (l *Logger) Success(msg string, kvs ...any) {
	l.emit(LevelSuccess, 3, msg, kvs)
}

func (l *Logger) Warn(msg string, kvs ...any) {
	l.emit(LevelWarn, 3, msg, kvs)
}

func (l *Logger) Error(msg string, kvs ...any) {
	l.emit(LevelError, 3, msg, kvs)
}

func (l *Logger) Fatal(msg string, kvs ...any) {
	l.emit(LevelFatal, 3, msg, kvs)
	os.Exit(1)
}

// --- format methods (printf-style) ---

func (l *Logger) Debugf(format string, args ...any) {
	l.emit(LevelDebug, 3, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Infof(format string, args ...any) {
	l.emit(LevelInfo, 3, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Successf(format string, args ...any) {
	l.emit(LevelSuccess, 3, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.emit(LevelWarn, 3, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.emit(LevelError, 3, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.emit(LevelFatal, 3, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

// --- println methods (space-separated args) ---

func (l *Logger) Debugln(args ...any) {
	l.emit(LevelDebug, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Infoln(args ...any) {
	l.emit(LevelInfo, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Successln(args ...any) {
	l.emit(LevelSuccess, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Warnln(args ...any) {
	l.emit(LevelWarn, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Errorln(args ...any) {
	l.emit(LevelError, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Fatalln(args ...any) {
	l.emit(LevelFatal, 3, fmt.Sprintln(args...), nil)
	os.Exit(1)
}

// --- print methods (no newline, no spaces — like fmt.Print) ---

func (l *Logger) Println(args ...any) {
	l.emit(LevelInfo, 3, fmt.Sprintln(args...), nil)
}

func (l *Logger) Printf(format string, args ...any) {
	l.emit(LevelInfo, 3, fmt.Sprintf(format, args...), nil)
}

// --- internal ---

// emit builds a Record, formats it, and writes it out.
// callerSkip is the number of stack frames to skip when resolving caller info.
func (l *Logger) emit(level Level, callerSkip int, msg string, kvs []any) {
	if level < l.cfg.MinLevel {
		return
	}

	rec := Record{
		Level:   level,
		Message: msg,
		Time:    time.Now(),
		Module:  l.cfg.Module,
		Fields:  makeFields(kvs),
	}

	if l.cfg.Caller || level >= LevelError {
		ci := resolveCaller(callerSkip)
		rec.File = ci.File
		rec.Line = ci.Line
		rec.Func = ci.Func
	}

	_ = l.writer.Write(l.fmt.Format(&rec))
	if l.fileW != nil {
		_ = l.fileW.Write(l.fileFmt.Format(&rec))
	}
}

// makeFields converts a flat key-value slice into []Field.
// Odd-length slices get a trailing "<MISSING>" value.
func makeFields(kvs []any) []Field {
	n := len(kvs)
	if n == 0 {
		return nil
	}
	fields := make([]Field, 0, (n+1)/2)
	for i := 0; i < n; i += 2 {
		key := ""
		if s, ok := kvs[i].(string); ok {
			key = s
		}
		var val any
		if i+1 < n {
			val = kvs[i+1]
		} else {
			val = "<MISSING>"
		}
		fields = append(fields, Field{Key: key, Value: val})
	}
	return fields
}
