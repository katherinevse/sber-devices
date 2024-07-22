package logger

import (
	"log/slog"
	"os"
)

func Configure() {
	level := slog.LevelInfo
	err := level.UnmarshalText([]byte(os.Getenv("LOG")))
	if err != nil {
		// TODO return error
		slog.Info("Undefined log level")
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
