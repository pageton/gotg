package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger returns a *zap.Logger that routes all gotd internal logs
// through this Logger instance, replacing JSON output with styled text.
func (l *Logger) ZapLogger() *zap.Logger {
	return zap.New(&zapBridge{logger: l})
}

type zapBridge struct {
	logger *Logger
	fields []Field
}

func (z *zapBridge) Enabled(lvl zapcore.Level) bool {
	return zapToLevel(lvl) >= z.logger.cfg.MinLevel
}

func (z *zapBridge) With(fields []zapcore.Field) zapcore.Core {
	extra := make([]Field, len(fields))
	for i, f := range fields {
		extra[i] = Field{Key: f.Key, Value: zapFieldValue(f)}
	}
	return &zapBridge{
		logger: z.logger,
		fields: append(z.fields, extra...),
	}
}

func (z *zapBridge) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if z.Enabled(entry.Level) {
		return ce.AddCore(entry, z)
	}
	return ce
}

func (z *zapBridge) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	lvl := zapToLevel(entry.Level)

	all := make([]any, 0, (len(z.fields)+len(fields))*2)
	for _, f := range z.fields {
		all = append(all, f.Key, f.Value)
	}
	for _, f := range fields {
		all = append(all, f.Key, zapFieldValue(f))
	}

	module := entry.LoggerName
	if module != "" {
		child := z.logger.WithModule(module)
		child.emit(lvl, 6, entry.Message, all)
	} else {
		z.logger.emit(lvl, 6, entry.Message, all)
	}
	return nil
}

func (z *zapBridge) Sync() error { return nil }

func zapToLevel(lvl zapcore.Level) Level {
	switch {
	case lvl <= zapcore.DebugLevel:
		return LevelDebug
	case lvl == zapcore.InfoLevel:
		return LevelInfo
	case lvl == zapcore.WarnLevel:
		return LevelWarn
	case lvl == zapcore.ErrorLevel:
		return LevelError
	default:
		return LevelFatal
	}
}

func zapFieldValue(f zapcore.Field) any {
	switch f.Type {
	case zapcore.StringType:
		return f.String
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return f.Integer
	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
		return uint64(f.Integer) //nolint:gosec // intentional; zapcore stores uint64 in int64 field
	case zapcore.Float64Type:
		return float64(f.Integer)
	case zapcore.Float32Type:
		return float32(f.Integer)
	case zapcore.BoolType:
		return f.Integer == 1
	case zapcore.DurationType:
		return f.Interface
	case zapcore.StringerType:
		if s, ok := f.Interface.(interface{ String() string }); ok {
			return s.String()
		}
		return f.Interface
	default:
		if f.Interface != nil {
			return f.Interface
		}
		return f.String
	}
}
