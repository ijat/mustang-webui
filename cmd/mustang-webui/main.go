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
	dev := flag.Bool("dev", false, "use system java + a locally built sidecar jar, and proxy the frontend to a Vite dev server")
	devFrontendURL := flag.String("dev-frontend", "http://localhost:5173", "Vite dev server URL, used only with --dev")
	port := flag.Int("port", 0, "port to listen on (0 = pick a free port)")
	noBrowser := flag.Bool("no-browser", false, "don't open a browser window on start")
	verbose := flag.Bool("verbose", false, "log structured diagnostics to stderr instead of the default quiet progress output")
	flag.Parse()

	configureLogging(*verbose)
	reporter := orchestrator.NewReporter(os.Stdout)

	fmt.Println("mustang-webui")

	if err := run(runOptions{
		dev:            *dev,
		devFrontendURL: *devFrontendURL,
		port:           *port,
		noBrowser:      *noBrowser,
		reporter:       reporter,
	}); err != nil {
		reporter.Blank()
		fmt.Println(orchestrator.ExplainError(err))
		slog.Error("mustang-webui exited with error", "error", err)
		os.Exit(1)
	}
}

// configureLogging sets up log/slog as the structured, --verbose-only
// diagnostic layer. The default (non-verbose) run relies entirely on the
// Reporter for user-facing output; slog stays quiet except for warnings
// and above, so the two never compete for the same lines.
func configureLogging(verbose bool) {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

type runOptions struct {
	dev            bool
	devFrontendURL string
	port           int
	noBrowser      bool
	reporter       *orchestrator.Reporter
}

func run(opts runOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cacheDir, err := appdir.CacheDir()
	if err != nil {
		return fmt.Errorf("resolving cache directory: %w", err)
	}
	slog.Debug("resolved cache directory", "dir", cacheDir)

	rt, err := orchestrator.ProvisionRuntime(ctx, cacheDir, opts.dev, opts.reporter)
	if err != nil {
		return fmt.Errorf("provisioning runtime: %w", err)
	}

	opts.reporter.Section("Starting…")

	sidecar, err := orchestrator.StartSidecar(ctx, rt, opts.reporter)
	if err != nil {
		return fmt.Errorf("starting sidecar: %w", err)
	}
	defer sidecar.Stop()

	handlerOpts := server.Options{Sidecar: sidecar}
	if opts.dev {
		handlerOpts.DevFrontendURL = opts.devFrontendURL
	}

	handler, err := server.New(handlerOpts)
	if err != nil {
		return fmt.Errorf("building server: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.port))
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	url := "http://" + listener.Addr().String()
	opts.reporter.Ok(fmt.Sprintf("Serving on %s", url))

	if !opts.noBrowser {
		openBrowser(url)
	}

	opts.reporter.Blank()
	fmt.Println("Opened in your browser. Leave this window open while you work.")
	fmt.Println("Press Ctrl+C to quit.")

	httpServer := &http.Server{Handler: handler}
	errCh := make(chan error, 1)
	go func() { errCh <- httpServer.Serve(listener) }()

	select {
	case <-ctx.Done():
		slog.Debug("shutting down")
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
