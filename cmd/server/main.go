package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/service"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
	"github.com/neizhmak/avito-review-service/internal/transport/rest"
)

func main() {
	dbConnStr := os.Getenv("DB_CONNECTION_STRING")
	if dbConnStr == "" {
		dbConnStr = "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Warn("failed to close db", "error", err)
		}
	}()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	// initialize storages
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)

	// initialize service
	prService := service.NewPRService(prStorage, userStorage, teamStorage, db)

	// initialize handler (HTTP)
	handler := rest.NewHandler(prService)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler.InitRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("starting server", "port", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
