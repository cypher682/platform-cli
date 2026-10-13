package logutil

import (
	"io"
	"log/slog"
	"os"
)

var Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

func EnableVerbose() {
	Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func Disable() {
	Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}
