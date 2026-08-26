package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/opstrack/internal/models"
)

// PostgresServiceRepository is a PostgreSQL-backed ServiceRepository.
type PostgresServiceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresServiceRepository(pool *pgxpool.Pool) *PostgresServiceRepository {
	return &PostgresServiceRepository{pool: pool}
}

func (r *PostgresServiceRepository) Create(ctx context.Context, s *models.Service) error {
	const q = `
		INSERT INTO services (name, description, owner, environment, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, s.Name, s.Description, s.Owner, s.Environment, s.Status).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *PostgresServiceRepository) Get(ctx context.Context, id int64) (*models.Service, error) {
	const q = `
		SELECT id, name, description, owner, environment, status, created_at, updated_at
		FROM services WHERE id = $1`
	var s models.Service
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Name, &s.Description, &s.Owner, &s.Environment, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresServiceRepository) List(ctx context.Context, filter ServiceFilter) ([]models.Service, error) {
	q := `SELECT id, name, description, owner, environment, status, created_at, updated_at FROM services WHERE 1=1`
	var args []any
	if filter.Status != "" {
		args = append(args, filter.Status)
		q += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if filter.Environment != "" {
		args = append(args, filter.Environment)
		q += fmt.Sprintf(" AND environment = $%d", len(args))
	}
	q += " ORDER BY id"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Owner, &s.Environment, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PostgresServiceRepository) Update(ctx context.Context, s *models.Service) error {
	const q = `
		UPDATE services SET name=$1, description=$2, owner=$3, environment=$4, status=$5, updated_at=now()
		WHERE id=$6
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q, s.Name, s.Description, s.Owner, s.Environment, s.Status, s.ID).Scan(&s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *PostgresServiceRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM services WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// PostgresIncidentRepository is a PostgreSQL-backed IncidentRepository.
type PostgresIncidentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresIncidentRepository(pool *pgxpool.Pool) *PostgresIncidentRepository {
	return &PostgresIncidentRepository{pool: pool}
}

func (r *PostgresIncidentRepository) Create(ctx context.Context, i *models.Incident) error {
	const q = `
		INSERT INTO incidents (service_id, title, description, severity, status, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, i.ServiceID, i.Title, i.Description, i.Severity, i.Status, i.AssignedTo).
		Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
}

func (r *PostgresIncidentRepository) Get(ctx context.Context, id int64) (*models.Incident, error) {
	const q = `
		SELECT id, service_id, title, description, severity, status, assigned_to, created_at, updated_at, resolved_at
		FROM incidents WHERE id = $1`
	var i models.Incident
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&i.ID, &i.ServiceID, &i.Title, &i.Description, &i.Severity, &i.Status, &i.AssignedTo,
		&i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *PostgresIncidentRepository) List(ctx context.Context, filter IncidentFilter) ([]models.Incident, error) {
	q := `SELECT id, service_id, title, description, severity, status, assigned_to, created_at, updated_at, resolved_at
		FROM incidents WHERE 1=1`
	var args []any
	var conds []string
	if filter.Status != "" {
		args = append(args, filter.Status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if filter.Severity != "" {
		args = append(args, filter.Severity)
		conds = append(conds, fmt.Sprintf("severity = $%d", len(args)))
	}
	if filter.ServiceID != 0 {
		args = append(args, filter.ServiceID)
		conds = append(conds, fmt.Sprintf("service_id = $%d", len(args)))
	}
	if len(conds) > 0 {
		q += " AND " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY id"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Incident
	for rows.Next() {
		var i models.Incident
		if err := rows.Scan(&i.ID, &i.ServiceID, &i.Title, &i.Description, &i.Severity, &i.Status, &i.AssignedTo,
			&i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *PostgresIncidentRepository) Update(ctx context.Context, i *models.Incident) error {
	const q = `
		UPDATE incidents
		SET service_id=$1, title=$2, description=$3, severity=$4, status=$5, assigned_to=$6, updated_at=now(),
		    resolved_at = CASE WHEN $5 = 'RESOLVED' AND resolved_at IS NULL THEN now() ELSE resolved_at END
		WHERE id=$7
		RETURNING updated_at, resolved_at`
	err := r.pool.QueryRow(ctx, q, i.ServiceID, i.Title, i.Description, i.Severity, i.Status, i.AssignedTo, i.ID).
		Scan(&i.UpdatedAt, &i.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *PostgresIncidentRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM incidents WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
