package models

import "time"

// ServiceStatus represents the operational status of a monitored service.
type ServiceStatus string

const (
	ServiceStatusActive      ServiceStatus = "ACTIVE"
	ServiceStatusDegraded    ServiceStatus = "DEGRADED"
	ServiceStatusDown        ServiceStatus = "DOWN"
	ServiceStatusMaintenance ServiceStatus = "MAINTENANCE"
)

func (s ServiceStatus) Valid() bool {
	switch s {
	case ServiceStatusActive, ServiceStatusDegraded, ServiceStatusDown, ServiceStatusMaintenance:
		return true
	}
	return false
}

// Service represents an application/service monitored by the operations team.
type Service struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Owner       string        `json:"owner"`
	Environment string        `json:"environment"`
	Status      ServiceStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ServiceInput is the payload accepted for creating/updating a Service.
type ServiceInput struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Owner       string        `json:"owner"`
	Environment string        `json:"environment"`
	Status      ServiceStatus `json:"status"`
}

// IncidentSeverity represents how severe an incident is.
type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "LOW"
	SeverityMedium   IncidentSeverity = "MEDIUM"
	SeverityHigh     IncidentSeverity = "HIGH"
	SeverityCritical IncidentSeverity = "CRITICAL"
)

func (s IncidentSeverity) Valid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	}
	return false
}

// IncidentStatus represents the lifecycle state of an incident.
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "OPEN"
	IncidentStatusInvestigating IncidentStatus = "INVESTIGATING"
	IncidentStatusMitigated     IncidentStatus = "MITIGATED"
	IncidentStatusResolved      IncidentStatus = "RESOLVED"
)

func (s IncidentStatus) Valid() bool {
	switch s {
	case IncidentStatusOpen, IncidentStatusInvestigating, IncidentStatusMitigated, IncidentStatusResolved:
		return true
	}
	return false
}

// allowedIncidentTransitions defines which status transitions are legal.
var allowedIncidentTransitions = map[IncidentStatus][]IncidentStatus{
	IncidentStatusOpen:          {IncidentStatusInvestigating, IncidentStatusResolved},
	IncidentStatusInvestigating: {IncidentStatusMitigated, IncidentStatusResolved},
	IncidentStatusMitigated:     {IncidentStatusResolved, IncidentStatusInvestigating},
	IncidentStatusResolved:      {},
}

// CanTransition reports whether moving from one incident status to another is allowed.
func CanTransition(from, to IncidentStatus) bool {
	if from == to {
		return true
	}
	for _, allowed := range allowedIncidentTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// Incident represents an operational incident linked to a Service.
type Incident struct {
	ID          int64            `json:"id"`
	ServiceID   int64            `json:"service_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	AssignedTo  string           `json:"assigned_to,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ResolvedAt  *time.Time       `json:"resolved_at,omitempty"`
}

// IncidentInput is the payload accepted for creating/updating an Incident.
type IncidentInput struct {
	ServiceID   int64            `json:"service_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	AssignedTo  string           `json:"assigned_to,omitempty"`
}
