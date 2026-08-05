// Command mustang-webui is the single entry point users run: it provisions
// the JVM sidecar (downloading it on first run, or reusing a local build
// in --dev), starts it, serves the frontend on loopback, and opens the
// user's browser. Nothing here ever touches a non-loopback address except
// the one-time asset download.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ijat/mustang-webui/internal/appdir"
	"github.com/ijat/mustang-webui/internal/orchestrator"
	"github.com/ijat/mustang-webui/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("mustang-webui exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dev := flag.Bool("dev", false, "use system java + a locally built sidecar jar, and proxy the frontend to a Vite dev server")
	devFrontendURL := flag.String("dev-frontend", "http://localhost:5173", "Vite dev server URL, used only with --dev")
	port := flag.Int("port", 0, "port to listen on (0 = pick a free port)")
	noBrowser := flag.Bool("no-browser", false, "don't open a browser window on start")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cacheDir, err := appdir.CacheDir()
	if err != nil {
		return fmt.Errorf("resolving cache directory: %w", err)
	}

	slog.Info("provisioning runtime", "dev", *dev, "cacheDir", cacheDir)
	rt, err := orchestrator.ProvisionRuntime(ctx, cacheDir, *dev)
	if err != nil {
		return fmt.Errorf("provisioning runtime: %w", err)
	}

	slog.Info("starting sidecar")
	sidecar, err := orchestrator.StartSidecar(ctx, rt)
	if err != nil {
		return fmt.Errorf("starting sidecar: %w", err)
	}
	defer sidecar.Stop()

	opts := server.Options{Sidecar: sidecar}
	if *dev {
		opts.DevFrontendURL = *devFrontendURL
	}

	handler, err := server.New(opts)
	if err != nil {
		return fmt.Errorf("building server: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	addr := listener.Addr().String()
	url := "http://" + addr
	slog.Info("serving", "url", url)

	if !*noBrowser {
		openBrowser(url)
	}

	httpServer := &http.Server{Handler: handler}
	errCh := make(chan error, 1)
	go func() { errCh <- httpServer.Serve(listener) }()

	select {
	case <-ctx.Done():
		slog.Info("shutting down")
		return httpServer.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		slog.Warn("could not open browser automatically", "url", url, "error", err)
	}
}
