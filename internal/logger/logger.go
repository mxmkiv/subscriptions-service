package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {

	/*
		check logger level
	*/

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(logHandler)
}
