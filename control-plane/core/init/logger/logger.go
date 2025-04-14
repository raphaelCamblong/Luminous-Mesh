package logger

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func NewLogger() *zap.Logger {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer l.Sync()
	logger = l
	return l
}

func L() *zap.Logger {
	return logger
}
