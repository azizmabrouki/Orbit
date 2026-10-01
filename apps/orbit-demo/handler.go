package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// app holds the state shared by the HTTP handlers.
type app struct {
	cfg          config
	version      string
	hostname     string
	logger       *slog.Logger
	metrics      *metrics
	shuttingDown atomic.Bool
}

func newApp(cfg config, version, hostname string, logger *slog.Logger) *app {
	return &app{
		cfg:      cfg,
		version:  version,
		hostname: hostname,
		logger:   logger,
		metrics:  newMetrics(version),
	}
}

// routeLabels maps every registered pattern to the value of the route label.
// Anything not listed here, including the catch-all, is labelled
// "unmatched", so the label never holds a raw path.
var routeLabels = map[string]string{
	"GET /{$}":     "/",
	"GET /healthz": "/healthz",
	"GET /readyz":  "/readyz",
	"GET /metrics": "/metrics",
	// Same paths with any other method: answered with 405.
	"/{$}":     "/",
	"/healthz": "/healthz",
	"/readyz":  "/readyz",
	"/metrics": "/metrics",
}

// quietPatterns are probes and scrapes, logged at debug instead of info.
var quietPatterns = map[string]bool{
	"GET /healthz": true,
	"GET /readyz":  true,
	"GET /metrics": true,
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.handleRoot)
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("GET /readyz", a.handleReadyz)
	mux.Handle("GET /metrics", promhttp.HandlerFor(a.metrics.registry, promhttp.HandlerOpts{}))
	// ServeMux answers a wrong method and an unknown path in plain text, and a
	// catch-all "/" would turn every wrong method into a 404. Registering each
	// path without a method keeps both answers JSON.
	for _, p := range []string{"/{$}", "/healthz", "/readyz", "/metrics"} {
		mux.HandleFunc(p, handleMethodNotAllowed)
	}
	mux.HandleFunc("/", handleNotFound)
	return a.instrument(mux)
}

func (a *app) handleRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		Service  string `json:"service"`
		Message  string `json:"message"`
		Version  string `json:"version"`
		Hostname string `json:"hostname"`
	}{"orbit-demo", a.cfg.Message, a.version, a.hostname})
}

// handleHealthz is liveness: it answers while the process runs and checks
// nothing else.
func (a *app) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReadyz is readiness. orbit-demo has no dependencies, so it is ready
// whenever the listener accepts requests, until shutdown starts.
func (a *app) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if a.shuttingDown.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "shutting down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func handleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", "GET, HEAD")
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// instrument records metrics and writes one access-log line per request.
func (a *app) instrument(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// For its internal redirects (unclean paths such as /a/../healthz),
		// the mux returns the target path instead of a pattern. Every route is
		// an exact path, so a real match always has route == path.
		_, pattern := mux.Handler(r)
		route, ok := routeLabels[pattern]
		if !ok || route != r.URL.Path {
			route = "unmatched"
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		mux.ServeHTTP(rec, r)
		elapsed := time.Since(start)

		method := methodLabel(r.Method)
		a.metrics.requests.WithLabelValues(method, route, strconv.Itoa(rec.status)).Inc()
		a.metrics.duration.WithLabelValues(method, route).Observe(elapsed.Seconds())

		level := slog.LevelInfo
		if quietPatterns[pattern] {
			level = slog.LevelDebug
		}
		a.logger.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("route", route),
			slog.Int("status", rec.status),
			slog.Float64("duration_ms", float64(elapsed.Microseconds())/1000),
		)
	})
}

// methodLabel bounds the method label: clients can send any method string.
func methodLabel(m string) string {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
		http.MethodConnect, http.MethodTrace:
		return m
	}
	return "OTHER"
}

// statusRecorder captures the status code written by the wrapped handler.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

type metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newMetrics(version string) *metrics {
	m := &metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests handled, by method, route and status code.",
		}, []string{"method", "route", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds, by method and route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
	}
	buildInfo := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "orbit_demo_build_info",
		Help:        "Build information of orbit-demo. Always 1.",
		ConstLabels: prometheus.Labels{"version": version},
	})
	buildInfo.Set(1)

	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.requests,
		m.duration,
		buildInfo,
	)
	return m
}
