// Package validation contains business-rule validation for OpsTrack domain
// objects, independent of HTTP or database concerns so it can be unit
// tested in isolation.
package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/example/opstrack/internal/models"
)

var (
	ErrNameRequired        = errors.New("name is required")
	ErrOwnerRequired       = errors.New("owner is required")
	ErrEnvironmentRequired = errors.New("environment is required")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrTitleRequired       = errors.New("title is required")
	ErrInvalidServiceID    = errors.New("service_id must be a positive integer")
	ErrInvalidSeverity     = errors.New("invalid severity")
	ErrInvalidTransition   = errors.New("invalid incident status transition")
)

// ValidateServiceInput validates fields for creating/updating a Service.
func ValidateServiceInput(in models.ServiceInput) error {
	var errs []string

	if strings.TrimSpace(in.Name) == "" {
		errs = append(errs, ErrNameRequired.Error())
	}
	if strings.TrimSpace(in.Owner) == "" {
		errs = append(errs, ErrOwnerRequired.Error())
	}
	if strings.TrimSpace(in.Environment) == "" {
		errs = append(errs, ErrEnvironmentRequired.Error())
	}
	if in.Status != "" && !in.Status.Valid() {
		errs = append(errs, ErrInvalidStatus.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// ValidateIncidentInput validates fields for creating/updating an Incident.
func ValidateIncidentInput(in models.IncidentInput) error {
	var errs []string

	if in.ServiceID <= 0 {
		errs = append(errs, ErrInvalidServiceID.Error())
	}
	if strings.TrimSpace(in.Title) == "" {
		errs = append(errs, ErrTitleRequired.Error())
	}
	if !in.Severity.Valid() {
		errs = append(errs, ErrInvalidSeverity.Error())
	}
	if in.Status != "" && !in.Status.Valid() {
		errs = append(errs, ErrInvalidStatus.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// ValidateIncidentTransition checks that moving from one incident status to
// another follows the allowed state machine.
func ValidateIncidentTransition(from, to models.IncidentStatus) error {
	if !models.CanTransition(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}
