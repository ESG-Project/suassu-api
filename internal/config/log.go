package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BuildLogger(cfg *Config) (*zap.Logger, error) {
	level := zap.InfoLevel
	switch cfg.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeDuration = zapcore.MillisDurationEncoder
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		level,
	)

	// Campos padrão
	logger := zap.New(core).With(
		zap.String("app", cfg.AppName),
		zap.String("env", cfg.AppEnv),
	)
	return logger, nil
}
