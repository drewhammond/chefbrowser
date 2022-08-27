package logging

import (
	"github.com/drewhammond/chefbrowser/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field zapcore.Field

type Logger struct {
	*zap.Logger
}

func New(config *config.Config) *Logger {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{config.Logging.Output}
	cfg.Encoding = config.Logging.Format

	// disable because it's not that useful outside of development
	cfg.DisableCaller = true

	// disable sampling and use ISO8601 timestamps
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	level, _ := zapcore.ParseLevel(config.Logging.Level)
	cfg.Level.SetLevel(level)
	cfg.Sampling = nil
	cfg.EncoderConfig = ec
	logger, _ := cfg.Build()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger) // flushes buffer, if any
	l := &Logger{logger}
	return l
}
