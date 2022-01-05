package logger

import (
	"log"

	"go.uber.org/zap"
)

var (
	logger       *zap.Logger
	compatLogger *log.Logger
)

func New(isProd bool) (*zap.Logger, error) {
	var zapConfig zap.Config

	if isProd {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

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
