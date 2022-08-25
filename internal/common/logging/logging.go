package logging

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New() *Logger {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = viper.GetStringSlice("logging.destination")

	// disable because it's not that useful outside of development
	cfg.DisableCaller = true

	// disable sampling and use ISO8601 timestamps
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	level, _ := zapcore.ParseLevel(viper.GetString("logging.level"))
	cfg.Level.SetLevel(level)
	cfg.Sampling = nil
	cfg.EncoderConfig = ec
	logger, _ := cfg.Build()
	defer logger.Sync() // flushes buffer, if any
	l := &Logger{logger.Named("chefbrowser")}
	return l
}
