package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"subscriptions-service/internal/config"
	httpHandler "subscriptions-service/internal/handler/http"
	"subscriptions-service/internal/repository/postgres"
	"subscriptions-service/internal/service"
)

// @title           Subscriptions Service API
// @version         1.0
// @description     A service for managing user subscriptions.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Logger
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log.Info("config loaded successfully")

	// Database
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN())
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	log.Info("database connection established")

	// Migrations
	m, err := migrate.New(
		"file://migrations",
		cfg.Database.DSN(),
	)
	if err != nil {
		log.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	log.Info("migrations applied successfully")

	// Initialize repository, service, handler and router
	repo := postgres.NewSubscriptionRepository(pool, log)
	svc := service.NewSubscriptionService(repo, log)
	h := httpHandler.NewHandler(svc, log)
	router := h.InitRoutes()

	// Server
	log.Info("starting server", "port", cfg.Server.Port)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("server exited properly")
}
