package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mxmkiv/subscriptions-service/internal/config"
	"github.com/mxmkiv/subscriptions-service/internal/connection"
	"github.com/mxmkiv/subscriptions-service/internal/handler"
	"github.com/mxmkiv/subscriptions-service/internal/logger"
	"github.com/mxmkiv/subscriptions-service/internal/repository"
	"github.com/mxmkiv/subscriptions-service/internal/service"
)

func main() {

	// tmp logger
	bootstrapLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// config
	config, err := config.NewConfig()
	if err != nil {
		bootstrapLogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// slog logger object
	logger := logger.New(config.Logger.Level)
	logger.Info("logger initialized", "level", config.Logger.Level)

	// db connection
	database, err := connection.NewPostgres(context.Background(), config.DB.DSN())
	if err != nil {
		logger.Error("failed to connect db", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// layers
	repo := repository.New(database)
	service := service.New(repo)
	handler := handler.New(service, logger)

	// routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscription", handler.Create)
	mux.HandleFunc("GET /subscription/{id}", handler.GetByID)
	mux.HandleFunc("PUT /subscription/{id}", handler.Update)
	mux.HandleFunc("DELETE /subscription/{id}", handler.Delete)
	mux.HandleFunc("GET /subscriptions", handler.List)
	mux.HandleFunc("GET /subscriptions/total", handler.SumByPeriod)

	// server
	logger.Info("server start", "addr", ":"+config.HTTP.Port)
	err = http.ListenAndServe(":"+config.HTTP.Port, mux)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
