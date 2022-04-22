//go:build windows

package server

import (
	"os"
	"syscall"

	"go.uber.org/zap"
)

func SendKillSignal(slog *zap.SugaredLogger) {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		slog.Errorw("failed to find windows process", "err", err)
	}

	err = p.Signal(syscall.SIGINT)
	if err != nil {
		slog.Errorw("failed to trigger windows signal", "err", err)
	}
}
