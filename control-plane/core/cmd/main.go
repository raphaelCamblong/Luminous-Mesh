package main

import (
	"os"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/logger"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/process"
	"go.uber.org/zap"
)

func main() {
	logger.NewLogger()
	defer handlePanic()

	proc := process.NewProcess()
	proc.Launch()
}

func handlePanic() {
	if r := recover(); r != nil {
		logger.L().Error("⚠️ Panic intercepted", zap.Any("panic", r))
		os.Exit(-1)
	}
}
