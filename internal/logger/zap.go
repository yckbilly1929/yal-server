package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger       *zap.Logger
	compatLogger *log.Logger
)

type LoggerOpt struct {
	Debug bool
	Color bool

	// TODO
	Encoding string
}

func New(opt LoggerOpt) (*zap.Logger, error) {
	var zapConfig zap.Config

	if opt.Debug {
		zapConfig = zap.NewDevelopmentConfig()
		if opt.Color {
			zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.DisableCaller = true
	}
	zapConfig.Encoding = "console"

	var err error
	logger, err = zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func Compat() (*log.Logger, error) {
	if logger == nil {
		panic("logger not init")
	}
	// TODO: init once
	if compatLogger == nil {
		var err error
		compatLogger, err = zap.NewStdLogAt(logger.Named("stdlib"), zap.DebugLevel)
		if err != nil {
			return nil, err
		}
	}

	return compatLogger, nil
}

func L() *zap.Logger {
	if logger == nil {
		panic("logger not init")
	}

	return logger
}

func S() *zap.SugaredLogger {
	if logger == nil {
		panic("logger not init")
	}

	return logger.Sugar()
}
