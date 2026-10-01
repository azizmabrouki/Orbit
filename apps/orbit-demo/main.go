// Command orbit-demo is Orbit's sample workload and the reference
// implementation of the Golden Path workload contract.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	// Restore default signal handling once shutdown starts, so a second
	// signal kills the process at once.
	context.AfterFunc(ctx, stop)
	code := run(ctx, os.Getenv, os.Stdout)
	stop()
	os.Exit(code)
}

// run starts the server and blocks until ctx is cancelled and shutdown
// completes. It returns the process exit code.
func run(ctx context.Context, getenv func(string) string, stdout io.Writer) int {
	cfg, err := loadConfig(getenv)
	if err != nil {
		logger := slog.New(slog.NewJSONHandler(stdout, nil))
		logger.Error("invalid configuration", "error", err.Error())
		return 1
	}
	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: cfg.level()}))

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		logger.Error("listen failed", "error", err.Error())
		return 1
	}

	a := newApp(cfg, version, hostname, logger)
	logger.Info("starting orbit-demo",
		"version", version,
		"addr", ln.Addr().String(),
		slog.Group("config",
			"port", cfg.Port,
			"log_level", cfg.LogLevel,
			"message", cfg.Message,
			"shutdown_drain", cfg.ShutdownDrain.String(),
			"shutdown_timeout", cfg.ShutdownTimeout.String(),
		),
	)

	if err := a.serve(ctx, ln, a.routes()); err != nil {
		logger.Error("server stopped with an error", "error", err.Error())
		return 1
	}
	return 0
}

// serve runs an HTTP server on ln until ctx is cancelled, then shuts down in
// three steps: readiness turns 503, the server keeps serving for the drain
// period so Kubernetes can remove the pod from its endpoints, and in-flight
// requests get up to the shutdown timeout to finish.
func (a *app) serve(ctx context.Context, ln net.Listener, h http.Handler) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(a.logger.Handler(), slog.LevelError),
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	a.shuttingDown.Store(true)
	// Close connections after their current request, so clients reconnect
	// and reach a pod that is still ready.
	srv.SetKeepAlivesEnabled(false)
	a.logger.Info("shutting down",
		"drain", a.cfg.ShutdownDrain.String(),
		"timeout", a.cfg.ShutdownTimeout.String(),
	)

	select {
	case <-time.After(a.cfg.ShutdownDrain):
	case err := <-errc:
		return fmt.Errorf("serve: %w", err)
	}

	a.logger.Info("drain finished, closing listener")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	a.logger.Info("shutdown complete")
	return nil
}
