package repository

import (
	"context"
	"sync"
	"time"

	"github.com/example/opstrack/internal/models"
)

// InMemoryServiceRepository is a thread-safe in-memory ServiceRepository,
// used for unit tests and local experimentation without PostgreSQL.
type InMemoryServiceRepository struct {
	mu     sync.RWMutex
	nextID int64
	data   map[int64]models.Service
}

func NewInMemoryServiceRepository() *InMemoryServiceRepository {
	return &InMemoryServiceRepository{data: make(map[int64]models.Service)}
}

func (r *InMemoryServiceRepository) Create(_ context.Context, s *models.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	s.ID = r.nextID
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	r.data[s.ID] = *s
	return nil
}

func (r *InMemoryServiceRepository) Get(_ context.Context, id int64) (*models.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &s, nil
}

func (r *InMemoryServiceRepository) List(_ context.Context, filter ServiceFilter) ([]models.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []models.Service
	for _, s := range r.data {
		if filter.Status != "" && string(s.Status) != filter.Status {
			continue
		}
		if filter.Environment != "" && s.Environment != filter.Environment {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *InMemoryServiceRepository) Update(_ context.Context, s *models.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.data[s.ID]
	if !ok {
		return ErrNotFound
	}
	s.CreatedAt = existing.CreatedAt
	s.UpdatedAt = time.Now().UTC()
	r.data[s.ID] = *s
	return nil
}

func (r *InMemoryServiceRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return ErrNotFound
	}
	delete(r.data, id)
	return nil
}

// InMemoryIncidentRepository is a thread-safe in-memory IncidentRepository.
type InMemoryIncidentRepository struct {
	mu     sync.RWMutex
	nextID int64
	data   map[int64]models.Incident
}

func NewInMemoryIncidentRepository() *InMemoryIncidentRepository {
	return &InMemoryIncidentRepository{data: make(map[int64]models.Incident)}
}

func (r *InMemoryIncidentRepository) Create(_ context.Context, i *models.Incident) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	i.ID = r.nextID
	now := time.Now().UTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	r.data[i.ID] = *i
	return nil
}

func (r *InMemoryIncidentRepository) Get(_ context.Context, id int64) (*models.Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &i, nil
}

func (r *InMemoryIncidentRepository) List(_ context.Context, filter IncidentFilter) ([]models.Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []models.Incident
	for _, i := range r.data {
		if filter.Status != "" && string(i.Status) != filter.Status {
			continue
		}
		if filter.Severity != "" && string(i.Severity) != filter.Severity {
			continue
		}
		if filter.ServiceID != 0 && i.ServiceID != filter.ServiceID {
			continue
		}
		out = append(out, i)
	}
	return out, nil
}

func (r *InMemoryIncidentRepository) Update(_ context.Context, i *models.Incident) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.data[i.ID]
	if !ok {
		return ErrNotFound
	}
	i.CreatedAt = existing.CreatedAt
	i.UpdatedAt = time.Now().UTC()
	if i.Status == models.IncidentStatusResolved && i.ResolvedAt == nil {
		now := time.Now().UTC()
		i.ResolvedAt = &now
	}
	r.data[i.ID] = *i
	return nil
}

func (r *InMemoryIncidentRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return ErrNotFound
	}
	delete(r.data, id)
	return nil
}
