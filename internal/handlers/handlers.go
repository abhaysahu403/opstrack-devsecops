// Package handlers implements the HTTP transport layer for OpsTrack,
// translating requests into service-layer calls and back into JSON.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/opstrack/internal/models"
	"github.com/example/opstrack/internal/repository"
	"github.com/example/opstrack/internal/service"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

// DBChecker abstracts a readiness check against the database.
type DBChecker func(ctx context.Context) error

type API struct {
	Services  *service.ServiceService
	Incidents *service.IncidentService
	DBCheck   DBChecker
}

func NewAPI(services *service.ServiceService, incidents *service.IncidentService, dbCheck DBChecker) *API {
	return &API{Services: services, Incidents: incidents, DBCheck: dbCheck}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// Routes registers all OpsTrack HTTP routes on the given mux using Go 1.22
// method-aware routing patterns.
func (a *API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", a.Info)
	mux.HandleFunc("GET /health", a.Health)
	mux.HandleFunc("GET /ready", a.Ready)

	mux.HandleFunc("POST /api/v1/services", a.CreateService)
	mux.HandleFunc("GET /api/v1/services", a.ListServices)
	mux.HandleFunc("GET /api/v1/services/{id}", a.GetService)
	mux.HandleFunc("PUT /api/v1/services/{id}", a.UpdateService)
	mux.HandleFunc("DELETE /api/v1/services/{id}", a.DeleteService)

	mux.HandleFunc("POST /api/v1/incidents", a.CreateIncident)
	mux.HandleFunc("GET /api/v1/incidents", a.ListIncidents)
	mux.HandleFunc("GET /api/v1/incidents/{id}", a.GetIncident)
	mux.HandleFunc("PUT /api/v1/incidents/{id}", a.UpdateIncident)
	mux.HandleFunc("DELETE /api/v1/incidents/{id}", a.DeleteIncident)
}

func (a *API) Info(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"application": "OpsTrack",
		"version":     "1.0.0",
		"status":      "running",
	})
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (a *API) Ready(w http.ResponseWriter, r *http.Request) {
	if a.DBCheck == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "READY"})
		return
	}
	if err := a.DBCheck(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "NOT_READY", "reason": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "READY"})
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func handleRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case strings.Contains(err.Error(), "validation:"):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ---- Services ----

func (a *API) CreateService(w http.ResponseWriter, r *http.Request) {
	var in models.ServiceInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := a.Services.Create(r.Context(), in)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, svc)
}

func (a *API) ListServices(w http.ResponseWriter, r *http.Request) {
	filter := repository.ServiceFilter{
		Status:      r.URL.Query().Get("status"),
		Environment: r.URL.Query().Get("environment"),
	}
	list, err := a.Services.List(r.Context(), filter)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *API) GetService(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	svc, err := a.Services.Get(r.Context(), id)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (a *API) UpdateService(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in models.ServiceInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := a.Services.Update(r.Context(), id, in)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (a *API) DeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := a.Services.Delete(r.Context(), id); err != nil {
		handleRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Incidents ----

func (a *API) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var in models.IncidentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	inc, err := a.Incidents.Create(r.Context(), in)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, inc)
}

func (a *API) ListIncidents(w http.ResponseWriter, r *http.Request) {
	filter := repository.IncidentFilter{
		Status:   r.URL.Query().Get("status"),
		Severity: r.URL.Query().Get("severity"),
	}
	if sid := r.URL.Query().Get("service_id"); sid != "" {
		if v, err := strconv.ParseInt(sid, 10, 64); err == nil {
			filter.ServiceID = v
		}
	}
	list, err := a.Incidents.List(r.Context(), filter)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *API) GetIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	inc, err := a.Incidents.Get(r.Context(), id)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inc)
}

func (a *API) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in models.IncidentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	inc, err := a.Incidents.Update(r.Context(), id, in)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inc)
}

func (a *API) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := a.Incidents.Delete(r.Context(), id); err != nil {
		handleRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
