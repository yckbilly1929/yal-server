package server

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/yckbilly1929/yal-server/internal/debounce"
	"github.com/yckbilly1929/yal-server/internal/logger"
	"github.com/yckbilly1929/yal-server/internal/middleware"
)

var (
	initOnce          sync.Once
	insecureTransport *http.Transport

	mu         sync.RWMutex
	activeConn = map[net.Conn]struct{}{}

	wsPort                   int
	wsConnectedPayload       = []byte("connected")
	wsReloadPayload          = []byte("reload")
	wsRefreshCssPayload      = []byte("refresh-css")
	wsRefreshCssPopupPayload = []byte("refresh-css-popup")

	defaultRootDir      = "dist"
	defaultRootFile     = "index.html"
	defaultCertFilePath = "./server.crt"
	defaultKeyFilePath  = "./server.key"

	headEnd     = []byte("</head>")
	bodyEnd     = []byte("</body>")
	replacePort = "{{port}}"
	//go:embed build/out.js
	injectedCode string
	//go:embed build/out.css
	injectedStyle string

	isRewatching atomic.Bool
)

func Run(sc ServeConfig) {
	// TODO: validate config
	if sc.Root == "" {
		sc.Root = defaultRootDir
	}
	if sc.File == "" {
		sc.File = defaultRootFile
	}
	if sc.HTTPS {
		// TODO: create fake cert, check has package.json / is yarn, set filePath
	}

	// init
	llog, err := logger.New(false)
	if err != nil {
		panic(err)
	}
	compatLog, err := logger.Compat()
	if err != nil {
		panic(err)
	}
	slog := llog.Sugar()

	// 1. serve
	r := chi.NewRouter()
	r.Use(middleware.Base())

	if sc.CORS {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{
				`https://*`,
				`http://*`,
			},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}

	for _, p := range sc.Proxy {
		r.HandleFunc(p.Prefix+"/*", reverseHandler(p))
	}

	rootDir := http.Dir(sc.Root)
	fs := http.FileServer(rootDir)

	rootHandler := serveFileContents(defaultRootFile, rootDir)

	if sc.Fallback {
		r.Get("/", rootHandler)
		r.Get("/*", middleware.Intercept404(fs, rootHandler))
	} else {
		r.Get("/", rootHandler)
		r.Get("/*", fs.ServeHTTP)
	}

	h := &http.Server{
		Addr:     fmt.Sprintf(":%d", sc.Port),
		Handler:  r,
		ErrorLog: compatLog,
	}

	// 2. wss
	wsServer := &http.Server{
		Addr: ":0", // random assign port later
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, _, _, err := ws.UpgradeHTTP(r, w)
			if err != nil {
				// TODO: handle error
				slog.Errorw("failed to UpgradeHTTP", "err", err)
				return
			}

			slog.Infow("ws connected", "addr", conn.RemoteAddr().String())
			mu.Lock()
			defer mu.Unlock()
			activeConn[conn] = struct{}{}

			// reader
			go func() {
				for {
					// TODO: detect disconnected and remove from map?
					msg, op, err := wsutil.ReadClientData(conn)
					slog.Debugw("ReadClientData", "msg", msg, "op", op)
					if err != nil {
						var closedError wsutil.ClosedError
						if closed := errors.As(err, &closedError); closed {
							// TODO: print non-standard error?
							// slog.Debugw("debug ClosedError", "code", closedError.Code, "reason", closedError.Reason)
						} else {
							slog.Errorw("failed to ReadClientData", "err", err)
						}
						break
					}
				}

				slog.Infow("ws disconnected", "addr", conn.RemoteAddr().String())
				mu.Lock()
				defer mu.Unlock()
				delete(activeConn, conn)
			}()

			err = wsutil.WriteServerMessage(conn, ws.OpText, wsConnectedPayload)
			if err != nil {
				// TODO: handle error
				slog.Errorw("failed to WriteServerMessage", "err", err)
			}
		}),
		ErrorLog: compatLog,
	}

	// 3. watch
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Fatal(err)
	}
	// TODO: remove ignored file patterns, recursive watch on different platforms
	err = watcher.Add(sc.Root)
	if err != nil {
		slog.Fatal(err)
	}

	// wait for shut down in a separate goroutine.
	errCh := make(chan error)
	go gracefullShutdown(h, wsServer, watcher, errCh)

	startApp(&sc, h, wsServer, watcher)

	if err := <-errCh; err != nil {
		slog.Errorw("could not gracefully shutdown server", "err", err)
		return
	}

	slog.Info("server stopped")
}

func startApp(sc *ServeConfig, httpServer *http.Server, wsServer *http.Server, watcher *fsnotify.Watcher) {
	slog := logger.S()
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		slog.Info("http server will start")
		var err error
		if sc.HTTPS {
			err = httpServer.ListenAndServeTLS(defaultCertFilePath, defaultKeyFilePath)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	g.Go(func() error {
		slog.Info("ws server will start")
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
		wsPort = listener.Addr().(*net.TCPAddr).Port
		if sc.HTTPS {
			err = wsServer.ServeTLS(listener, defaultCertFilePath, defaultKeyFilePath)
		} else {
			err = wsServer.Serve(listener)
		}

		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	g.Go(func() error {
		slog.Info("file watcher will start")
		reloadDebouncer := debounce.New(time.Millisecond * 100)
		refreshCssDebouncer := debounce.New(time.Millisecond * 100)
		reloadFunc := genNotifyWsFunc(wsReloadPayload)
		refreshCssFunc := genNotifyWsFunc(wsRefreshCssPayload)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					// TODO: already closed
					return nil
				}

				// TODO: rewatch directory
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if event.Name == sc.Root {
						err := tryRewatch(sc, watcher)
						if err != nil {
							// TODO: error handling on watch target missing
							return err
						}
					}
					continue
				}

				// TODO: reload depends on event? e.g. separate css and others
				if event.Op&fsnotify.Write == fsnotify.Write {
					isRefreshCss := strings.Contains(event.Name, ".css")
					if isRefreshCss {
						refreshCssDebouncer(refreshCssFunc)
					} else {
						reloadDebouncer(reloadFunc)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					// TODO: already closed
					return nil
				}

				// TODO: error handling
				slog.Errorw("failed on watcher", "err", err)
			}
		}
	})

	go func() {
		// The context is closed if both servers finish, or one of them
		// errors out, in which case we want to close the other and return.
		<-ctx.Done()

		// TODO: not available on windows server
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			slog.Errorw("failed to trigger syscall kill", "err", err)
		}
	}()

	err := g.Wait()
	if err != nil {
		slog.Fatalw("could not start server", "err", err)
	}
}

func gracefullShutdown(httpServer *http.Server, wsServer *http.Server, watcher *fsnotify.Watcher, errCh chan<- error) {
	slog := logger.S()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM) // from terminal or k8s

	<-quit
	slog.Info("server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		httpServer.SetKeepAlivesEnabled(false)
		return httpServer.Shutdown(ctx)
	})
	g.Go(func() error {
		wsServer.SetKeepAlivesEnabled(false)
		// TODO: close all connections?
		return wsServer.Shutdown(ctx)
	})
	g.Go(func() error {
		return watcher.Close()
	})

	errCh <- g.Wait()
}
