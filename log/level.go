package log

type Level uint8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
	LevelFatal
	// LevelOff disables all log output. Used by Nop().
	LevelOff
)

func (l Level) Tag() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelSuccess:
		return "OK"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "???"
	}
}

func (l Level) Icon() string {
	switch l {
	case LevelDebug:
		return "●"
	case LevelInfo:
		return "▸"
	case LevelSuccess:
		return "✔"
	case LevelWarn:
		return "▲"
	case LevelError:
		return "✖"
	case LevelFatal:
		return "✖"
	default:
		return "·"
	}
}

const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cCyan   = "\033[36m"
	cWhite  = "\033[37m"
)

func (l Level) Color() string {
	switch l {
	case LevelDebug:
		return cDim + cCyan
	case LevelInfo:
		return cBlue
	case LevelSuccess:
		return cGreen
	case LevelWarn:
		return cYellow
	case LevelError:
		return cBold + cRed
	case LevelFatal:
		return cBold + cRed
	default:
		return cWhite
	}
}
