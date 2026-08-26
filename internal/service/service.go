// Package service contains OpsTrack business logic, sitting between HTTP
// handlers and the repository layer. It applies validation and enforces
// domain rules such as incident status transitions.
package service

import (
	"context"
	"fmt"

	"github.com/example/opstrack/internal/models"
	"github.com/example/opstrack/internal/repository"
	"github.com/example/opstrack/internal/validation"
)

type ServiceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) Create(ctx context.Context, in models.ServiceInput) (*models.Service, error) {
	if err := validation.ValidateServiceInput(in); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	status := in.Status
	if status == "" {
		status = models.ServiceStatusActive
	}
	svc := &models.Service{
		Name:        in.Name,
		Description: in.Description,
		Owner:       in.Owner,
		Environment: in.Environment,
		Status:      status,
	}
	if err := s.repo.Create(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) Get(ctx context.Context, id int64) (*models.Service, error) {
	return s.repo.Get(ctx, id)
}

func (s *ServiceService) List(ctx context.Context, filter repository.ServiceFilter) ([]models.Service, error) {
	return s.repo.List(ctx, filter)
}

func (s *ServiceService) Update(ctx context.Context, id int64, in models.ServiceInput) (*models.Service, error) {
	if err := validation.ValidateServiceInput(in); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.Owner = in.Owner
	existing.Environment = in.Environment
	if in.Status != "" {
		existing.Status = in.Status
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *ServiceService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

type IncidentService struct {
	repo        repository.IncidentRepository
	serviceRepo repository.ServiceRepository
}

func NewIncidentService(repo repository.IncidentRepository, serviceRepo repository.ServiceRepository) *IncidentService {
	return &IncidentService{repo: repo, serviceRepo: serviceRepo}
}

func (s *IncidentService) Create(ctx context.Context, in models.IncidentInput) (*models.Incident, error) {
	if err := validation.ValidateIncidentInput(in); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	if s.serviceRepo != nil {
		if _, err := s.serviceRepo.Get(ctx, in.ServiceID); err != nil {
			return nil, fmt.Errorf("service_id %d: %w", in.ServiceID, err)
		}
	}
	status := in.Status
	if status == "" {
		status = models.IncidentStatusOpen
	}
	inc := &models.Incident{
		ServiceID:   in.ServiceID,
		Title:       in.Title,
		Description: in.Description,
		Severity:    in.Severity,
		Status:      status,
		AssignedTo:  in.AssignedTo,
	}
	if err := s.repo.Create(ctx, inc); err != nil {
		return nil, err
	}
	return inc, nil
}

func (s *IncidentService) Get(ctx context.Context, id int64) (*models.Incident, error) {
	return s.repo.Get(ctx, id)
}

func (s *IncidentService) List(ctx context.Context, filter repository.IncidentFilter) ([]models.Incident, error) {
	return s.repo.List(ctx, filter)
}

func (s *IncidentService) Update(ctx context.Context, id int64, in models.IncidentInput) (*models.Incident, error) {
	if err := validation.ValidateIncidentInput(in); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Status != "" && in.Status != existing.Status {
		if err := validation.ValidateIncidentTransition(existing.Status, in.Status); err != nil {
			return nil, err
		}
		existing.Status = in.Status
	}
	existing.ServiceID = in.ServiceID
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Severity = in.Severity
	existing.AssignedTo = in.AssignedTo

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *IncidentService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
