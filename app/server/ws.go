package server

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/yckbilly1929/yal-server/internal/logger"
)

func genNotifyWsFunc(payload []byte) func() {
	slog := logger.S()

	return func() {
		mu.RLock()
		defer mu.RUnlock()
		for conn := range activeConn {
			err := wsutil.WriteServerMessage(conn, ws.OpText, payload)
			if err != nil {
				// TODO: handle error
				slog.Errorw("failed to WriteServerMessage", "err", err)
			}
		}
	}
}
