package main

import (
	"log/slog"
	"os"

	"github.com/mxmkiv/subscriptions-service/internal/config"
	"github.com/mxmkiv/subscriptions-service/internal/connection"
	"github.com/mxmkiv/subscriptions-service/internal/logger"
)

func main() {

	bootstrapLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := config.NewConfig()
	if err != nil {
		bootstrapLogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := logger.New()

	// db connection
	database, err := connection.NewPostgres(config.DB.DSN())
	if err != nil {
		logger.Error("failed to connect db", "error", err)
		os.Exit(1)
	}
	defer database.Close()

}
