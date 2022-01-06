package server

import (
	"crypto/tls"
	"errors"
	"io/fs"
	"net/http"
	"os"
)

type ServeConfig struct {
	Port               uint   `json:"port"`
	Root               string `json:"root"`
	File               string `json:"file"`
	CORS               bool   `json:"cors"`
	HTTPS              bool   `json:"https"`
	HistoryApiFallback bool   `json:"historyApiFallback"`

	Proxy  []ProxyMiddleware `json:"proxy"`
	Server Server            `json:"server"`

	// TODO
	Host   string `json:"host"`
	IsYarn bool   `json:"isYarn"`

	Internal internalConfig `json:"-"`
}

type Server struct {
	// file name only, not path
	Cert string `json:"cert"`
	Key  string `json:"key"`

	// logger
	Debug bool `json:"debug"`
	Color bool `json:"color"`
}

type ProxyMiddleware struct {
	Prefix       string `json:"prefix"`
	Target       string `json:"target"`
	ChangeOrigin bool   `json:"changeOrigin"`
	Secure       bool   `json:"secure"`
}

type internalConfig struct {
	CertPath string `json:"-"`
	KeyPath  string `json:"-"`
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

func getCacheDir(sc ServeConfig) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// TODO: check if package.json
	isNodePkg := true
	_, err = os.Stat(cwd + "/" + nodePkgFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		isNodePkg = false
	}

	if !isNodePkg {
		// TODO
		return ".cache/yalive-server", nil
	}
	if sc.IsYarn {
		// TODO
		return ".yarn/.cache/yalive-server", nil
	}

	return "node_modules/.cache/yalive-server", nil
}
