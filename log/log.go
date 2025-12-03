package log

import (
	"os"

	"github.com/ooqls/getset/log/api/v1/gen"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var level zap.AtomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
var encodingConfig *zapcore.EncoderConfig = &zapcore.EncoderConfig{
	MessageKey:     "log",
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "name",
	CallerKey:      "caller",
	FunctionKey:    "func",
	StacktraceKey:  "stack",
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeName:     zapcore.FullNameEncoder,
}
var encoder zapcore.Encoder = zapcore.NewJSONEncoder(*encodingConfig)

var core zapcore.Core = zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)
var l *zap.Logger = zap.New(core)

var cfg *zap.Config = &zap.Config{
	Level:            level,
	OutputPaths:      []string{"stdout"},
	ErrorOutputPaths: []string{"stderr"},
	Encoding:         "json",
	EncoderConfig:    *encodingConfig,
}

type LogLevel string

func SetLogLevel(lvl gen.LevelEnum) {
	switch lvl {
	case gen.DEBUG:
		level.SetLevel(zap.DebugLevel)
	case gen.INFO:
		level.SetLevel(zap.InfoLevel)
	case gen.WARNING:
		level.SetLevel(zap.WarnLevel)
	case gen.ERROR:
		level.SetLevel(zap.ErrorLevel)
	default:
		level.SetLevel(zap.DebugLevel)
	}
}

func NewLogger(name string) *zap.Logger {
	return l.Named(name)
}

func GetLevel() gen.LevelEnum {
	switch level.Level() {
	case zap.DebugLevel:
		return gen.DEBUG
	case zap.InfoLevel:
		return gen.INFO
	case zap.WarnLevel:
		return gen.WARNING
	case zap.ErrorLevel:
		return gen.ERROR
	default:
		return gen.DEBUG
	}
}

func SetFormat(f gen.FormatEnum) error {
	switch f {
	case gen.JSON:
		encoder = zapcore.NewJSONEncoder(*encodingConfig)
	case gen.TEXT:
		encoder = zapcore.NewConsoleEncoder(*encodingConfig)
	default:
		encoder = zapcore.NewJSONEncoder(*encodingConfig)
	}
	core = zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)
	l = zap.New(core)

	return nil
}

func GetFormat() gen.FormatEnum {
	switch cfg.Encoding {
	case "json":
		return gen.JSON
	case "console":
		return gen.TEXT
	default:
		return gen.JSON
	}
}
