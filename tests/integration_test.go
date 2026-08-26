//go:build integration

// Package tests contains integration tests that run against a real
// PostgreSQL instance. They are excluded from the default `go test ./...`
// run via the "integration" build tag so unit tests stay fast and
// dependency-free. Run with:
//
//	go test -tags=integration ./tests/...
//
// A PostgreSQL instance must be reachable using the DB_* environment
// variables (see .env.example). In CI this is provided as a GitHub
// Actions service container.
package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/example/opstrack/internal/config"
	"github.com/example/opstrack/internal/database"
	"github.com/example/opstrack/internal/models"
	"github.com/example/opstrack/internal/repository"
)

func testConfig() config.Config {
	cfg := config.Load()
	if v := os.Getenv("TEST_DB_NAME"); v != "" {
		cfg.DBName = v
	}
	return cfg
}

func TestPostgresServiceAndIncidentLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := testConfig()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := database.Ping(ctx, pool); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "../migrations"
	}
	if err := database.RunMigrations(ctx, pool, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	svcRepo := repository.NewPostgresServiceRepository(pool)
	incRepo := repository.NewPostgresIncidentRepository(pool)

	svc := &models.Service{
		Name: "Integration Test Service", Owner: "qa-team", Environment: "test", Status: models.ServiceStatusActive,
	}
	if err := svcRepo.Create(ctx, svc); err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	if svc.ID == 0 {
		t.Fatal("expected service ID to be assigned")
	}

	inc := &models.Incident{
		ServiceID: svc.ID, Title: "Integration Test Incident", Severity: models.SeverityMedium, Status: models.IncidentStatusOpen,
	}
	if err := incRepo.Create(ctx, inc); err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	fetched, err := incRepo.Get(ctx, inc.ID)
	if err != nil {
		t.Fatalf("failed to fetch incident: %v", err)
	}
	if fetched.Title != "Integration Test Incident" {
		t.Fatalf("unexpected incident title: %s", fetched.Title)
	}

	inc.Status = models.IncidentStatusResolved
	if err := incRepo.Update(ctx, inc); err != nil {
		t.Fatalf("failed to update incident: %v", err)
	}
	if inc.ResolvedAt == nil {
		t.Fatal("expected resolved_at to be set after resolving")
	}

	if err := incRepo.Delete(ctx, inc.ID); err != nil {
		t.Fatalf("failed to delete incident: %v", err)
	}
	if err := svcRepo.Delete(ctx, svc.ID); err != nil {
		t.Fatalf("failed to delete service: %v", err)
	}
}
