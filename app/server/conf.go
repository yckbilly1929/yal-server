package server

import (
	"crypto/tls"
	"net/http"
)

type ServeConfig struct {
	Port     uint   `json:"port"`
	Root     string `json:"root"`
	File     string `json:"file"`
	CORS     bool   `json:"cors"`
	HTTPS    bool   `json:"https"`
	Fallback bool   `json:"fallback"`

	Proxy []ProxyMiddleware `json:"proxy"`

	// TODO
	Host   string `json:"host"`
	IsDir  bool   `json:"isDir"`
	IsYarn bool   `json:"isYarn"`
}

type ProxyMiddleware struct {
	Prefix       string `json:"prefix"`
	Target       string `json:"target"`
	ChangeOrigin bool   `json:"changeOrigin"`
	Secure       bool   `json:"secure"`
}

func getInsecureTransport() http.RoundTripper {
	initOnce.Do(func() {
		insecureTransport = http.DefaultTransport.(*http.Transport).Clone()
		insecureTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	})

	return insecureTransport
}
