// Package repository defines persistence interfaces for OpsTrack domain
// objects and provides both a PostgreSQL-backed implementation and an
// in-memory implementation used for fast unit testing.
package repository

import (
	"context"
	"errors"

	"github.com/example/opstrack/internal/models"
)

var ErrNotFound = errors.New("resource not found")

type ServiceFilter struct {
	Status      string
	Environment string
}

type IncidentFilter struct {
	Status    string
	Severity  string
	ServiceID int64
}

// ServiceRepository persists Service entities.
type ServiceRepository interface {
	Create(ctx context.Context, s *models.Service) error
	Get(ctx context.Context, id int64) (*models.Service, error)
	List(ctx context.Context, filter ServiceFilter) ([]models.Service, error)
	Update(ctx context.Context, s *models.Service) error
	Delete(ctx context.Context, id int64) error
}

// IncidentRepository persists Incident entities.
type IncidentRepository interface {
	Create(ctx context.Context, i *models.Incident) error
	Get(ctx context.Context, id int64) (*models.Incident, error)
	List(ctx context.Context, filter IncidentFilter) ([]models.Incident, error)
	Update(ctx context.Context, i *models.Incident) error
	Delete(ctx context.Context, id int64) error
}
