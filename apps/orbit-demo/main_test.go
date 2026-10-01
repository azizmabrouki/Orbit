package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunInvalidConfig(t *testing.T) {
	out := &syncBuffer{}
	code := run(context.Background(), env(map[string]string{"PORT": "eighty"}), out)
	if code != 1 {
		t.Errorf("exit code %d, want 1", code)
	}
	lines := out.lines(t)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1:\n%s", len(lines), out)
	}
	if lines[0]["level"] != "ERROR" || !strings.Contains(lines[0]["error"].(string), "PORT") {
		t.Errorf("unexpected line %v", lines[0])
	}
}

func listen(t *testing.T) (net.Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln, "http://" + ln.Addr().String()
}

func TestServeGracefulShutdown(t *testing.T) {
	a, logs := newTestApp(t, slog.LevelInfo)
	a.cfg.ShutdownDrain = 2 * time.Second
	a.cfg.ShutdownTimeout = time.Second
	ln, base := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.serve(ctx, ln, a.routes()) }()

	get := func(path string) int {
		t.Helper()
		resp, err := http.Get(base + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}

	if s := get("/readyz"); s != 200 {
		t.Fatalf("/readyz before shutdown: %d, want 200", s)
	}

	cancel()
	// Wait for the "shutting down" line, then probe during the drain.
	deadline := time.Now().Add(time.Second)
	for !strings.Contains(logs.String(), `"msg":"shutting down"`) {
		if time.Now().After(deadline) {
			t.Fatal("no shutting down log line")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if s := get("/readyz"); s != 503 {
		t.Errorf("/readyz during drain: %d, want 503", s)
	}
	if s := get("/healthz"); s != 200 {
		t.Errorf("/healthz during drain: %d, want 200", s)
	}
	if s := get("/"); s != 200 {
		t.Errorf("/ during drain: %d, want 200", s)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not return")
	}
	if !strings.Contains(logs.String(), `"msg":"shutdown complete"`) {
		t.Errorf("no shutdown complete line:\n%s", logs)
	}
	if _, err := http.Get(base + "/healthz"); err == nil {
		t.Error("listener still accepting after shutdown")
	}
}

func TestServeShutdownTimeout(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	a.cfg.ShutdownDrain = 0
	a.cfg.ShutdownTimeout = 100 * time.Millisecond
	ln, base := listen(t)

	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.serve(ctx, ln, slow) }()
	go func() {
		if resp, err := http.Get(base + "/"); err == nil {
			resp.Body.Close()
		}
	}()
	<-started
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("serve returned %v, want a deadline error", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not return")
	}
}

// TestRunLifecycle runs the whole program with a context that is already
// cancelled: it starts, logs its configuration, shuts down and exits 0.
func TestRunLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := &syncBuffer{}
	code := run(ctx, env(map[string]string{
		"PORT": freePort(t), "SHUTDOWN_DRAIN": "0s", "MESSAGE": "hi",
	}), out)
	if code != 0 {
		t.Fatalf("exit code %d:\n%s", code, out)
	}

	lines := out.lines(t)
	var msgs []string
	for _, l := range lines {
		msgs = append(msgs, l["msg"].(string))
	}
	want := []string{"starting orbit-demo", "shutting down", "drain finished, closing listener", "shutdown complete"}
	if strings.Join(msgs, "|") != strings.Join(want, "|") {
		t.Fatalf("log messages %q, want %q", msgs, want)
	}
	cfg, _ := lines[0]["config"].(map[string]any)
	if lines[0]["version"] != version || cfg["message"] != "hi" ||
		cfg["shutdown_drain"] != "0s" || cfg["log_level"] != "info" {
		t.Errorf("startup line %v", lines[0])
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	return port
}
