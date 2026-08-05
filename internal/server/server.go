// Package server wires the loopback-only HTTP server that serves the
// frontend and proxies API calls to the JVM sidecar. It never binds
// anything but 127.0.0.1 — the app has no business talking to the network
// once its own assets are provisioned.
package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ijat/mustang-webui/internal/orchestrator"
	"github.com/ijat/mustang-webui/web"
)

// Options configures the server.
type Options struct {
	Sidecar *orchestrator.Sidecar

	// DevFrontendURL, if set, proxies "/" to a running Vite dev server
	// (e.g. http://localhost:5173) instead of serving the embedded build,
	// so the frontend gets hot reload during development.
	DevFrontendURL string
}

// New builds the top-level HTTP handler.
func New(opts Options) (http.Handler, error) {
	mux := http.NewServeMux()

	apiProxy, err := newSidecarProxy(opts.Sidecar)
	if err != nil {
		return nil, err
	}
	mux.Handle("/api/", apiProxy)

	frontend, err := newFrontendHandler(opts.DevFrontendURL)
	if err != nil {
		return nil, err
	}
	mux.Handle("/", frontend)

	return mux, nil
}

func newSidecarProxy(sc *orchestrator.Sidecar) (http.Handler, error) {
	target, err := url.Parse(sc.BaseURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("Authorization", "Bearer "+sc.Token)
	}
	proxy.ErrorLog = slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)

	return proxy, nil
}

func newFrontendHandler(devFrontendURL string) (http.Handler, error) {
	if devFrontendURL != "" {
		target, err := url.Parse(devFrontendURL)
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(target), nil
	}

	sub, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(sub)), nil
}
