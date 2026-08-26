package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/example/opstrack/internal/handlers"
	"github.com/example/opstrack/internal/models"
	"github.com/example/opstrack/internal/repository"
	"github.com/example/opstrack/internal/service"
)

func newTestAPI() *handlers.API {
	svcRepo := repository.NewInMemoryServiceRepository()
	incRepo := repository.NewInMemoryIncidentRepository()
	svcSvc := service.NewServiceService(svcRepo)
	incSvc := service.NewIncidentService(incRepo, svcRepo)
	return handlers.NewAPI(svcSvc, incSvc, func(ctx context.Context) error { return nil })
}

func newTestServer() *httptest.Server {
	api := newTestAPI()
	mux := http.NewServeMux()
	api.Routes(mux)
	return httptest.NewServer(mux)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["status"] != "UP" {
		t.Fatalf("expected status UP, got %s", body["status"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ready")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestInfoEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["application"] != "OpsTrack" {
		t.Fatalf("expected application OpsTrack, got %s", body["application"])
	}
}

func TestCreateAndGetService(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	payload := models.ServiceInput{Name: "Payment Service", Owner: "payments-team", Environment: "production"}
	b, _ := json.Marshal(payload)

	resp, err := http.Post(srv.URL+"/api/v1/services", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var created models.Service
	_ = json.NewDecoder(resp.Body).Decode(&created)
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	getResp, err := http.Get(srv.URL + "/api/v1/services/" + strconv.FormatInt(created.ID, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}
}

func TestCreateServiceInvalidPayload(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b, _ := json.Marshal(models.ServiceInput{})
	resp, err := http.Post(srv.URL+"/api/v1/services", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetServiceNotFound(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/services/999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateIncidentAgainstMissingService(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	payload := models.IncidentInput{ServiceID: 123, Title: "X", Severity: models.SeverityHigh}
	b, _ := json.Marshal(payload)

	resp, err := http.Post(srv.URL+"/api/v1/incidents", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 404 or 400 for missing service, got %d", resp.StatusCode)
	}
}
