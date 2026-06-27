package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mxmkiv/subscriptions-service/docs"
	"github.com/mxmkiv/subscriptions-service/internal/config"
	"github.com/mxmkiv/subscriptions-service/internal/connection"
	"github.com/mxmkiv/subscriptions-service/internal/handler"
	"github.com/mxmkiv/subscriptions-service/internal/logger"
	"github.com/mxmkiv/subscriptions-service/internal/middleware"
	"github.com/mxmkiv/subscriptions-service/internal/repository"
	"github.com/mxmkiv/subscriptions-service/internal/service"
	"github.com/mxmkiv/subscriptions-service/migration"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title       Subscriptions Service API
// @version     1.0
// @description REST API for managing user subscriptions
// @host        localhost:8888
// @BasePath    /
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

	// goose migration
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}

	goose.SetBaseFS(migration.FS)

	if err := goose.Up(database, "."); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// layers
	repo := repository.New(database)
	service := service.New(repo)
	handler := handler.New(service, logger)

	// routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", handler.Create)
	mux.HandleFunc("GET /subscriptions/{id}", handler.GetByID)
	mux.HandleFunc("PUT /subscriptions/{id}", handler.Update)
	mux.HandleFunc("DELETE /subscriptions/{id}", handler.Delete)
	mux.HandleFunc("GET /subscriptions", handler.List)
	mux.HandleFunc("GET /subscriptions/total", handler.SumByPeriod)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// log all requests, disable if log_level higher then level info
	loggedMux := middleware.LoggingRequests(logger, mux)

	// server
	logger.Info("server start", "addr", ":"+config.HTTP.Port)
	err = http.ListenAndServe(":"+config.HTTP.Port, loggedMux)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
