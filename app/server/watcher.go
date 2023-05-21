package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/yckbilly1929/yalive-server/internal/logger"
)

func serveFileContents(sc ServeConfig, filePath string, dir http.Dir) http.HandlerFunc {
	slog := logger.S()

	return func(w http.ResponseWriter, r *http.Request) {
		// Restrict only to instances where the browser is looking for an HTML file
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		// TODO: mutex vs atomic?
		for {
			if isRewatching.Load() {
				continue
			} else {
				break
			}
		}

		// Open the file and return its contents using http.ServeContent
		f, err := dir.Open(filePath)
		if err != nil {
			slog.Errorw("directory not found", "err", err, "filePath", filePath)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		fInfo, err := f.Stat()
		if err != nil {
			slog.Errorw("file not found", "err", err, "filePath", filePath)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(f)
		var res []byte

		// TODO: inject wss script at the end of </head> or </body>
		port := strconv.FormatInt(int64(wsPort), 10)
		rawStyle := "<style>" + injectedStyle + "</style>"
		rawScript := "<script>" + injectedCode + "</script>"
		rawScript = strings.Replace(rawScript, replaceLocalAddress, sc.LocalAddress, 1)
		rawScript = strings.Replace(rawScript, replaceNetworkAddress, sc.NetworkAddress, 1)
		rawScript = strings.Replace(rawScript, replacePort, port, 1)
		neededContent := append([]byte(rawStyle), []byte(rawScript)...)
		if bytes.Contains(buf.Bytes(), headEnd) {
			res = bytes.Replace(buf.Bytes(), headEnd, append(neededContent, headEnd...), 1)
		} else if bytes.Contains(buf.Bytes(), bodyEnd) {
			res = bytes.Replace(buf.Bytes(), bodyEnd, append(neededContent, bodyEnd...), 1)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, fInfo.Name(), fInfo.ModTime(), bytes.NewReader(res))
	}
}

func tryRewatch(sc *ServeConfig, watcher *fsnotify.Watcher) error {
	isRewatching.Store(true)

	err := watcher.Remove(sc.Root)
	if err != nil {
		// TODO: error handling for permission?
		// logger.S().Debugw("failed to remove", "err", err)
	}

	c := 0
	for {
		if c >= 50 {
			return fmt.Errorf("failed to rewatch")
		}
		_, err := os.Stat(sc.Root + "/" + defaultRootFile)
		if err != nil {
			// not yet ready
			if os.IsExist(err) {
				return err
			}
			time.Sleep(200 * time.Millisecond)
			c++
			continue
		}

		break
	}

	err = watcher.Add(sc.Root)
	if err != nil {
		return err
	}

	isRewatching.Store(false)
	return nil
}
