package application

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(level, encoding string) (*zap.SugaredLogger, error) {
	logConf := zap.NewProductionConfig()
	if strings.ToLower(level) == "debug" {
		logConf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	if encoding != "" {
		logConf.Encoding = encoding
	}
	logConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := logConf.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
