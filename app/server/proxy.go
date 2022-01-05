package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/yckbilly1929/yal-server/internal/logger"
)

func reverseHandler(p ProxyMiddleware) func(w http.ResponseWriter, r *http.Request) {
	slog := logger.S()

	return func(w http.ResponseWriter, r *http.Request) {
		proxyServer := r.Host
		if p.Target != "" {
			proxyServer = p.Target
		}

		// TODO: parse req to decide proxy server
		proxyURL, err := url.Parse(proxyServer)
		if err != nil {
			// TODO
			slog.Errorw("failed to parse proxy server", "err", err)
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(proxyURL)

		if !p.Secure {
			reverseProxy.Transport = getInsecureTransport()
		}
		if p.ChangeOrigin {
			r.Host = proxyURL.Host
			r.Header.Set("Host", proxyURL.Host)
		}

		reverseProxy.ServeHTTP(w, r)
	}
}
