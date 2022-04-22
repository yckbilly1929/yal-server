//go:build darwin || linux

package server

import (
	"syscall"

	"go.uber.org/zap"
)

func SendKillSignal(slog *zap.SugaredLogger) {
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		slog.Errorw("failed to trigger syscall kill", "err", err)
	}
}
