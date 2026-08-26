package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/opstrack/internal/models"
	"github.com/example/opstrack/internal/repository"
	"github.com/example/opstrack/internal/service"
)

func newServices() (*service.ServiceService, *service.IncidentService) {
	svcRepo := repository.NewInMemoryServiceRepository()
	incRepo := repository.NewInMemoryIncidentRepository()
	return service.NewServiceService(svcRepo), service.NewIncidentService(incRepo, svcRepo)
}

func TestServiceService_CreateAndGet(t *testing.T) {
	svcSvc, _ := newServices()
	ctx := context.Background()

	created, err := svcSvc.Create(ctx, models.ServiceInput{
		Name: "Payment Service", Owner: "payments-team", Environment: "production",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected assigned ID, got 0")
	}
	if created.Status != models.ServiceStatusActive {
		t.Fatalf("expected default status ACTIVE, got %s", created.Status)
	}

	fetched, err := svcSvc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching: %v", err)
	}
	if fetched.Name != "Payment Service" {
		t.Fatalf("expected name Payment Service, got %s", fetched.Name)
	}
}

func TestServiceService_CreateInvalid(t *testing.T) {
	svcSvc, _ := newServices()
	_, err := svcSvc.Create(context.Background(), models.ServiceInput{})
	if err == nil {
		t.Fatal("expected validation error for empty input, got nil")
	}
}

func TestServiceService_DeleteNotFound(t *testing.T) {
	svcSvc, _ := newServices()
	err := svcSvc.Delete(context.Background(), 999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestIncidentService_CreateRequiresExistingService(t *testing.T) {
	svcSvc, incSvc := newServices()
	ctx := context.Background()

	_, err := incSvc.Create(ctx, models.IncidentInput{
		ServiceID: 42, Title: "Something broke", Severity: models.SeverityHigh,
	})
	if err == nil {
		t.Fatal("expected error creating incident against non-existent service")
	}

	svc, err := svcSvc.Create(ctx, models.ServiceInput{Name: "Order Service", Owner: "team", Environment: "prod"})
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}

	inc, err := incSvc.Create(ctx, models.IncidentInput{
		ServiceID: svc.ID, Title: "Order latency", Severity: models.SeverityMedium,
	})
	if err != nil {
		t.Fatalf("unexpected error creating incident: %v", err)
	}
	if inc.Status != models.IncidentStatusOpen {
		t.Fatalf("expected default status OPEN, got %s", inc.Status)
	}
}

func TestIncidentService_UpdateEnforcesTransitions(t *testing.T) {
	svcSvc, incSvc := newServices()
	ctx := context.Background()

	svc, _ := svcSvc.Create(ctx, models.ServiceInput{Name: "Order Service", Owner: "team", Environment: "prod"})
	inc, _ := incSvc.Create(ctx, models.IncidentInput{
		ServiceID: svc.ID, Title: "Order latency", Severity: models.SeverityMedium,
	})

	// OPEN -> RESOLVED is allowed.
	resolved, err := incSvc.Update(ctx, inc.ID, models.IncidentInput{
		ServiceID: svc.ID, Title: inc.Title, Severity: inc.Severity, Status: models.IncidentStatusResolved,
	})
	if err != nil {
		t.Fatalf("unexpected error transitioning to RESOLVED: %v", err)
	}
	if resolved.Status != models.IncidentStatusResolved {
		t.Fatalf("expected RESOLVED status, got %s", resolved.Status)
	}

	// RESOLVED -> OPEN is not allowed.
	_, err = incSvc.Update(ctx, inc.ID, models.IncidentInput{
		ServiceID: svc.ID, Title: inc.Title, Severity: inc.Severity, Status: models.IncidentStatusOpen,
	})
	if err == nil {
		t.Fatal("expected error transitioning RESOLVED -> OPEN, got nil")
	}
}

func TestIncidentService_ListFilters(t *testing.T) {
	svcSvc, incSvc := newServices()
	ctx := context.Background()

	svc, _ := svcSvc.Create(ctx, models.ServiceInput{Name: "Order Service", Owner: "team", Environment: "prod"})
	_, _ = incSvc.Create(ctx, models.IncidentInput{ServiceID: svc.ID, Title: "A", Severity: models.SeverityHigh})
	_, _ = incSvc.Create(ctx, models.IncidentInput{ServiceID: svc.ID, Title: "B", Severity: models.SeverityLow})

	highOnly, err := incSvc.List(ctx, repository.IncidentFilter{Severity: string(models.SeverityHigh)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(highOnly) != 1 {
		t.Fatalf("expected 1 high severity incident, got %d", len(highOnly))
	}
}
