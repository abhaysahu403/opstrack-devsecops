package validation

import (
	"testing"

	"github.com/example/opstrack/internal/models"
)

func TestValidateServiceInput(t *testing.T) {
	cases := []struct {
		name    string
		input   models.ServiceInput
		wantErr bool
	}{
		{
			name: "valid",
			input: models.ServiceInput{
				Name: "Payment Service", Owner: "payments-team", Environment: "production",
				Status: models.ServiceStatusActive,
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			input:   models.ServiceInput{Owner: "team", Environment: "production"},
			wantErr: true,
		},
		{
			name:    "missing owner",
			input:   models.ServiceInput{Name: "Service", Environment: "production"},
			wantErr: true,
		},
		{
			name:    "missing environment",
			input:   models.ServiceInput{Name: "Service", Owner: "team"},
			wantErr: true,
		},
		{
			name:    "invalid status",
			input:   models.ServiceInput{Name: "Service", Owner: "team", Environment: "prod", Status: "BOGUS"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateServiceInput(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateServiceInput() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateIncidentInput(t *testing.T) {
	cases := []struct {
		name    string
		input   models.IncidentInput
		wantErr bool
	}{
		{
			name:    "valid",
			input:   models.IncidentInput{ServiceID: 1, Title: "Latency spike", Severity: models.SeverityHigh},
			wantErr: false,
		},
		{
			name:    "missing service id",
			input:   models.IncidentInput{Title: "Latency spike", Severity: models.SeverityHigh},
			wantErr: true,
		},
		{
			name:    "missing title",
			input:   models.IncidentInput{ServiceID: 1, Severity: models.SeverityHigh},
			wantErr: true,
		},
		{
			name:    "invalid severity",
			input:   models.IncidentInput{ServiceID: 1, Title: "X", Severity: "BOGUS"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIncidentInput(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateIncidentInput() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateIncidentTransition(t *testing.T) {
	cases := []struct {
		name    string
		from    models.IncidentStatus
		to      models.IncidentStatus
		wantErr bool
	}{
		{"open to investigating", models.IncidentStatusOpen, models.IncidentStatusInvestigating, false},
		{"open to resolved", models.IncidentStatusOpen, models.IncidentStatusResolved, false},
		{"investigating to mitigated", models.IncidentStatusInvestigating, models.IncidentStatusMitigated, false},
		{"resolved to open (illegal)", models.IncidentStatusResolved, models.IncidentStatusOpen, true},
		{"mitigated to investigating (reopen)", models.IncidentStatusMitigated, models.IncidentStatusInvestigating, false},
		{"same status is a no-op", models.IncidentStatusOpen, models.IncidentStatusOpen, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIncidentTransition(tc.from, tc.to)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateIncidentTransition(%s -> %s) error = %v, wantErr %v", tc.from, tc.to, err, tc.wantErr)
			}
		})
	}
}
