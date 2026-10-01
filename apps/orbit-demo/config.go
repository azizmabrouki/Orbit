package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// config is the effective configuration, read from environment variables.
type config struct {
	Port            int
	LogLevel        string
	Message         string
	ShutdownDrain   time.Duration
	ShutdownTimeout time.Duration
}

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// loadConfig reads the configuration through getenv (os.Getenv in
// production). Unset or empty variables take their default. Every invalid
// value is reported in the returned error.
func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		Port:            8080,
		LogLevel:        "info",
		Message:         "Hello from Orbit",
		ShutdownDrain:   5 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	var errs []error

	if v := getenv("PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			errs = append(errs, fmt.Errorf("PORT %q: must be an integer between 1 and 65535", v))
		} else {
			cfg.Port = port
		}
	}

	if v := getenv("LOG_LEVEL"); v != "" {
		if _, ok := logLevels[v]; !ok {
			errs = append(errs, fmt.Errorf("LOG_LEVEL %q: must be debug, info, warn or error", v))
		} else {
			cfg.LogLevel = v
		}
	}

	if v := getenv("MESSAGE"); v != "" {
		cfg.Message = v
	}

	if v := getenv("SHUTDOWN_DRAIN"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < 0 {
			errs = append(errs, fmt.Errorf("SHUTDOWN_DRAIN %q: must be a duration of 0 or more, such as 5s", v))
		} else {
			cfg.ShutdownDrain = d
		}
	}

	if v := getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT %q: must be a positive duration, such as 10s", v))
		} else {
			cfg.ShutdownTimeout = d
		}
	}

	return cfg, errors.Join(errs...)
}

func (c config) level() slog.Level {
	return logLevels[c.LogLevel]
}
