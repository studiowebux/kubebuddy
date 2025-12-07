package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
)

type serviceRepo struct {
	db *sql.DB
}

func (r *serviceRepo) Create(ctx context.Context, service *domain.Service) error {
	minSpecJSON, err := json.Marshal(service.MinSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal min_spec: %w", err)
	}

	maxSpecJSON, err := json.Marshal(service.MaxSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal max_spec: %w", err)
	}

	placementJSON, err := json.Marshal(service.Placement)
	if err != nil {
		return fmt.Errorf("failed to marshal placement: %w", err)
	}

	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO services (id, name, min_spec, max_spec, placement, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, service.ID, service.Name, string(minSpecJSON), string(maxSpecJSON),
	   string(placementJSON), service.CreatedAt, service.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

func (r *serviceRepo) Get(ctx context.Context, id string) (*domain.Service, error) {
	var service domain.Service
	var minSpecJSON, maxSpecJSON, placementJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, min_spec, max_spec, placement, created_at, updated_at
		FROM services
		WHERE id = ?
	`, id).Scan(&service.ID, &service.Name, &minSpecJSON, &maxSpecJSON,
		&placementJSON, &service.CreatedAt, &service.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	if err := json.Unmarshal([]byte(minSpecJSON), &service.MinSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal min_spec: %w", err)
	}

	if err := json.Unmarshal([]byte(maxSpecJSON), &service.MaxSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal max_spec: %w", err)
	}

	if err := json.Unmarshal([]byte(placementJSON), &service.Placement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal placement: %w", err)
	}

	return &service, nil
}

func (r *serviceRepo) GetByName(ctx context.Context, name string) (*domain.Service, error) {
	var service domain.Service
	var minSpecJSON, maxSpecJSON, placementJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, min_spec, max_spec, placement, created_at, updated_at
		FROM services
		WHERE name = ?
	`, name).Scan(&service.ID, &service.Name, &minSpecJSON, &maxSpecJSON,
		&placementJSON, &service.CreatedAt, &service.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil for upsert logic
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service by name: %w", err)
	}

	if err := json.Unmarshal([]byte(minSpecJSON), &service.MinSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal min_spec: %w", err)
	}

	if err := json.Unmarshal([]byte(maxSpecJSON), &service.MaxSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal max_spec: %w", err)
	}

	if err := json.Unmarshal([]byte(placementJSON), &service.Placement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal placement: %w", err)
	}

	return &service, nil
}

func (r *serviceRepo) List(ctx context.Context) ([]*domain.Service, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, min_spec, max_spec, placement, created_at, updated_at
		FROM services
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	services := make([]*domain.Service, 0)
	for rows.Next() {
		var service domain.Service
		var minSpecJSON, maxSpecJSON, placementJSON string

		err := rows.Scan(&service.ID, &service.Name, &minSpecJSON, &maxSpecJSON,
			&placementJSON, &service.CreatedAt, &service.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		if err := json.Unmarshal([]byte(minSpecJSON), &service.MinSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal min_spec: %w", err)
		}

		if err := json.Unmarshal([]byte(maxSpecJSON), &service.MaxSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal max_spec: %w", err)
		}

		if err := json.Unmarshal([]byte(placementJSON), &service.Placement); err != nil {
			return nil, fmt.Errorf("failed to unmarshal placement: %w", err)
		}

		services = append(services, &service)
	}

	return services, nil
}

func (r *serviceRepo) Update(ctx context.Context, service *domain.Service) error {
	minSpecJSON, err := json.Marshal(service.MinSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal min_spec: %w", err)
	}

	maxSpecJSON, err := json.Marshal(service.MaxSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal max_spec: %w", err)
	}

	placementJSON, err := json.Marshal(service.Placement)
	if err != nil {
		return fmt.Errorf("failed to marshal placement: %w", err)
	}

	service.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, `
		UPDATE services
		SET name = ?, min_spec = ?, max_spec = ?, placement = ?, updated_at = ?
		WHERE id = ?
	`, service.Name, string(minSpecJSON), string(maxSpecJSON),
	   string(placementJSON), service.UpdatedAt, service.ID)

	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}

func (r *serviceRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM services WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}
