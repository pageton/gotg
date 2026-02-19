package log

// Config controls logger behavior.
type Config struct {
	// MinLevel is the minimum severity to emit. Messages below this are discarded.
	MinLevel Level
	// Debug enables debug-level output (convenience alias for MinLevel = LevelDebug).
	Debug bool
	// Timestamp enables time prefix on each log line.
	Timestamp bool
	// Color enables ANSI color output. Disable for file/structured output.
	Color bool
	// Caller appends file:line to log lines.
	Caller bool
	// FuncName appends the function name alongside caller info.
	FuncName bool
	// Module sets a default module tag (e.g. "dispatcher", "adapter").
	Module string
	// LogFile enables file logging. When set, logs go to both console and
	// this file path. File output is always plain text (no ANSI colors).
	LogFile string
	// RotateMaxSize enables log rotation when > 0. Bytes per file before rotating.
	RotateMaxSize int64
	// RotateMaxBackups is the number of rotated files to keep (default 3).
	RotateMaxBackups int
}

// DefaultConfig returns a Config suitable for development.
func DefaultConfig() Config {
	return Config{
		MinLevel:  LevelInfo,
		Timestamp: true,
		Color:     true,
		Caller:    false,
	}
}
