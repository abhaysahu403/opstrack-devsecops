// Command server starts the OpsTrack HTTP API.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/opstrack/internal/config"
	"github.com/example/opstrack/internal/database"
	"github.com/example/opstrack/internal/handlers"
	"github.com/example/opstrack/internal/middleware"
	"github.com/example/opstrack/internal/repository"
	"github.com/example/opstrack/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer pool.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := database.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	serviceRepo := repository.NewPostgresServiceRepository(pool)
	incidentRepo := repository.NewPostgresIncidentRepository(pool)

	serviceSvc := service.NewServiceService(serviceRepo)
	incidentSvc := service.NewIncidentService(incidentRepo, serviceRepo)

	api := handlers.NewAPI(serviceSvc, incidentSvc, func(ctx context.Context) error {
		return database.Ping(ctx, pool)
	})

	mux := http.NewServeMux()
	api.Routes(mux)

	handler := middleware.Chain(mux, middleware.Recovery, middleware.Logging, middleware.JSONContentType)

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("OpsTrack listening on :%s (env=%s)", cfg.ServerPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("shutdown complete")
}
