package main

import (
	"strings"
	"testing"
	"time"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(env(nil))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	want := config{
		Port:            8080,
		LogLevel:        "info",
		Message:         "Hello from Orbit",
		ShutdownDrain:   5 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	if cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	cfg, err := loadConfig(env(map[string]string{
		"PORT":             "9090",
		"LOG_LEVEL":        "debug",
		"MESSAGE":          "Hello from dev",
		"SHUTDOWN_DRAIN":   "0s",
		"SHUTDOWN_TIMEOUT": "1m30s",
	}))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	want := config{
		Port:            9090,
		LogLevel:        "debug",
		Message:         "Hello from dev",
		ShutdownDrain:   0,
		ShutdownTimeout: 90 * time.Second,
	}
	if cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadConfigLogLevels(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "error"} {
		cfg, err := loadConfig(env(map[string]string{"LOG_LEVEL": lvl}))
		if err != nil {
			t.Errorf("LOG_LEVEL=%s: %v", lvl, err)
			continue
		}
		if got := cfg.level().String(); !strings.EqualFold(got, lvl) {
			t.Errorf("LOG_LEVEL=%s: level %s", lvl, got)
		}
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	tests := []struct {
		name, value string
	}{
		{"PORT", "http"},
		{"PORT", "0"},
		{"PORT", "65536"},
		{"PORT", "-1"},
		{"LOG_LEVEL", "verbose"},
		{"LOG_LEVEL", "INFO"},
		{"SHUTDOWN_DRAIN", "5"},
		{"SHUTDOWN_DRAIN", "-1s"},
		{"SHUTDOWN_TIMEOUT", "ten"},
		{"SHUTDOWN_TIMEOUT", "0s"},
		{"SHUTDOWN_TIMEOUT", "-5s"},
	}
	for _, tt := range tests {
		t.Run(tt.name+"="+tt.value, func(t *testing.T) {
			_, err := loadConfig(env(map[string]string{tt.name: tt.value}))
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.name) {
				t.Errorf("error %q does not name %s", err, tt.name)
			}
		})
	}
}

func TestLoadConfigReportsEveryInvalidValue(t *testing.T) {
	_, err := loadConfig(env(map[string]string{"PORT": "x", "LOG_LEVEL": "y"}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"PORT", "LOG_LEVEL"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %s", err, name)
		}
	}
}
