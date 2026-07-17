// Package logger provides a single structured logger for all VOT Tradings
// Go services, built on the standard library's slog.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON structured logger at the given level ("debug", "info",
// "warn", "error"). Unrecognized levels fall back to "info".
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
