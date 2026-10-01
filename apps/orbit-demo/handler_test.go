package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// syncBuffer is a bytes.Buffer safe for the server goroutines and the test to
// share.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// lines decodes every log line, failing the test if one is not JSON.
func (b *syncBuffer) lines(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	sc := bufio.NewScanner(strings.NewReader(b.String()))
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("log line is not JSON: %q: %v", sc.Text(), err)
		}
		out = append(out, m)
	}
	return out
}

func newTestApp(t *testing.T, level slog.Level) (*app, *syncBuffer) {
	t.Helper()
	cfg, err := loadConfig(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	logs := &syncBuffer{}
	logger := slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: level}))
	return newApp(cfg, "test-version", "test-host", logger), logs
}

func do(h http.Handler, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestEndpoints(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	h := a.routes()

	tests := []struct {
		method, target string
		status         int
		body           map[string]string
	}{
		{"GET", "/", 200, map[string]string{
			"service": "orbit-demo", "message": "Hello from Orbit",
			"version": "test-version", "hostname": "test-host",
		}},
		{"GET", "/healthz", 200, map[string]string{"status": "ok"}},
		{"GET", "/readyz", 200, map[string]string{"status": "ready"}},
		{"GET", "/nope", 404, map[string]string{"error": "not found"}},
		{"GET", "/healthz/extra", 404, map[string]string{"error": "not found"}},
		{"POST", "/nope", 404, map[string]string{"error": "not found"}},
		{"POST", "/", 405, map[string]string{"error": "method not allowed"}},
		{"POST", "/healthz", 405, map[string]string{"error": "method not allowed"}},
		{"DELETE", "/readyz", 405, map[string]string{"error": "method not allowed"}},
		{"PUT", "/metrics", 405, map[string]string{"error": "method not allowed"}},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			rec := do(h, tt.method, tt.target)
			if rec.Code != tt.status {
				t.Errorf("status %d, want %d", rec.Code, tt.status)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type %q, want application/json", ct)
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not JSON: %q", rec.Body.String())
			}
			if len(body) != len(tt.body) {
				t.Errorf("body %v, want %v", body, tt.body)
			}
			for k, v := range tt.body {
				if body[k] != v {
					t.Errorf("body[%q] = %q, want %q", k, body[k], v)
				}
			}
			if tt.status == 405 && rec.Header().Get("Allow") != "GET, HEAD" {
				t.Errorf("Allow %q, want \"GET, HEAD\"", rec.Header().Get("Allow"))
			}
		})
	}
}

func TestMessageOverride(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	a.cfg.Message = "Changed through Git"
	rec := do(a.routes(), "GET", "/")
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["message"] != "Changed through Git" {
		t.Errorf("message %q", body["message"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	h := a.routes()
	do(h, "GET", "/")
	rec := do(h, "GET", "/metrics")
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type %q, want Prometheus text format", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`orbit_demo_build_info{version="test-version"} 1`,
		`http_requests_total{code="200",method="GET",route="/"} 1`,
		`http_request_duration_seconds_bucket{method="GET",route="/",le="0.005"}`,
		"go_goroutines ",
		"process_resident_memory_bytes ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics does not contain %q", want)
		}
	}
}

func TestReadyzDuringShutdown(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	h := a.routes()
	a.shuttingDown.Store(true)

	rec := do(h, "GET", "/readyz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/readyz status %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"shutting down"`) {
		t.Errorf("/readyz body %q", rec.Body.String())
	}
	if rec := do(h, "GET", "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("/healthz status %d, want 200", rec.Code)
	}
}

func TestRouteLabelUsesPattern(t *testing.T) {
	a, _ := newTestApp(t, slog.LevelInfo)
	h := a.routes()
	do(h, "GET", "/healthz")
	do(h, "POST", "/healthz")
	do(h, "GET", "/users/42")
	do(h, "GET", "/users/43")
	do(h, "GET", "/x/../healthz") // the mux redirects this to the clean path
	do(h, "BREW", "/coffee")

	counts := []struct {
		method, route, code string
		want                float64
	}{
		{"GET", "/healthz", "200", 1},
		{"POST", "/healthz", "405", 1},
		{"GET", "unmatched", "404", 2},
		{"GET", "unmatched", "307", 1},
		{"OTHER", "unmatched", "404", 1},
	}
	for _, c := range counts {
		got := testutil.ToFloat64(a.metrics.requests.WithLabelValues(c.method, c.route, c.code))
		if got != c.want {
			t.Errorf("http_requests_total{method=%q,route=%q,code=%q} = %v, want %v",
				c.method, c.route, c.code, got, c.want)
		}
	}

	families, err := a.metrics.registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"/": true, "/healthz": true, "/readyz": true, "/metrics": true, "unmatched": true}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "route" && !allowed[l.GetValue()] {
					t.Errorf("%s has route=%q, which is not a registered pattern", f.GetName(), l.GetValue())
				}
			}
		}
	}
}

func TestAccessLog(t *testing.T) {
	a, logs := newTestApp(t, slog.LevelInfo)
	h := a.routes()
	do(h, "GET", "/")
	do(h, "GET", "/missing")
	do(h, "GET", "/healthz")
	do(h, "GET", "/readyz")
	do(h, "GET", "/metrics")

	lines := logs.lines(t)
	if len(lines) != 2 {
		t.Fatalf("got %d log lines at info, want 2 (probes and scrapes not logged):\n%s", len(lines), logs)
	}
	want := []map[string]any{
		{"level": "INFO", "msg": "request", "method": "GET", "path": "/", "route": "/", "status": float64(200)},
		{"level": "INFO", "msg": "request", "method": "GET", "path": "/missing", "route": "unmatched", "status": float64(404)},
	}
	for i, line := range lines {
		for k, v := range want[i] {
			if line[k] != v {
				t.Errorf("line %d: %s = %v, want %v", i, k, line[k], v)
			}
		}
		if _, ok := line["time"].(string); !ok {
			t.Errorf("line %d: no time field", i)
		}
		if _, ok := line["duration_ms"].(float64); !ok {
			t.Errorf("line %d: duration_ms missing or not a number", i)
		}
	}
}

func TestAccessLogProbesAtDebug(t *testing.T) {
	a, logs := newTestApp(t, slog.LevelDebug)
	h := a.routes()
	do(h, "GET", "/healthz")
	do(h, "GET", "/readyz")
	do(h, "GET", "/metrics")

	lines := logs.lines(t)
	if len(lines) != 3 {
		t.Fatalf("got %d log lines at debug, want 3:\n%s", len(lines), logs)
	}
	for i, line := range lines {
		if line["level"] != "DEBUG" {
			t.Errorf("line %d: level %v, want DEBUG", i, line["level"])
		}
	}
}
