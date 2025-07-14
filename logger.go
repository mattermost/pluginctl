package pluginctl

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// Logger is the global logger instance.
var Logger *slog.Logger

// InitLogger initializes the global logger.
func InitLogger() {
	// Create a tint handler for colorized output
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
		AddSource:  false,
		NoColor:    false,
	})

	Logger = slog.New(handler)
}

// SetLogLevel sets the minimum logging level.
func SetLogLevel(level slog.Level) {
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
		AddSource:  false,
		NoColor:    false,
	})

	Logger = slog.New(handler)
}
