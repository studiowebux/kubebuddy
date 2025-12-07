package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type componentRepo struct {
	db *sql.DB
}

func (r *componentRepo) Create(ctx context.Context, component *domain.Component) error {
	specsJSON, err := json.Marshal(component.Specs)
	if err != nil {
		return fmt.Errorf("failed to marshal specs: %w", err)
	}

	query := `
		INSERT INTO components (id, name, type, manufacturer, model, specs, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		component.ID,
		component.Name,
		component.Type,
		component.Manufacturer,
		component.Model,
		specsJSON,
		component.Notes,
		component.CreatedAt,
		component.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	return nil
}

func (r *componentRepo) Get(ctx context.Context, id string) (*domain.Component, error) {
	query := `
		SELECT id, name, type, manufacturer, model, specs, notes, created_at, updated_at
		FROM components
		WHERE id = ?
	`

	var component domain.Component
	var specsJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&component.ID,
		&component.Name,
		&component.Type,
		&component.Manufacturer,
		&component.Model,
		&specsJSON,
		&component.Notes,
		&component.CreatedAt,
		&component.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("component not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	if err := json.Unmarshal([]byte(specsJSON), &component.Specs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal specs: %w", err)
	}

	return &component, nil
}

func (r *componentRepo) GetByManufacturerAndModel(ctx context.Context, manufacturer, model string) (*domain.Component, error) {
	query := `
		SELECT id, name, type, manufacturer, model, specs, notes, created_at, updated_at
		FROM components
		WHERE manufacturer = ? AND model = ?
	`

	var component domain.Component
	var specsJSON string

	err := r.db.QueryRowContext(ctx, query, manufacturer, model).Scan(
		&component.ID,
		&component.Name,
		&component.Type,
		&component.Manufacturer,
		&component.Model,
		&specsJSON,
		&component.Notes,
		&component.CreatedAt,
		&component.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	if err := json.Unmarshal([]byte(specsJSON), &component.Specs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal specs: %w", err)
	}

	return &component, nil
}

func (r *componentRepo) List(ctx context.Context, filters storage.ComponentFilters) ([]*domain.Component, error) {
	query := "SELECT id, name, type, manufacturer, model, specs, notes, created_at, updated_at FROM components WHERE 1=1"
	args := []interface{}{}

	if filters.Type != "" {
		query += " AND type = ?"
		args = append(args, filters.Type)
	}

	if filters.Manufacturer != "" {
		query += " AND manufacturer = ?"
		args = append(args, filters.Manufacturer)
	}

	query += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}
	defer rows.Close()

	var components []*domain.Component
	for rows.Next() {
		var component domain.Component
		var specsJSON string

		err := rows.Scan(
			&component.ID,
			&component.Name,
			&component.Type,
			&component.Manufacturer,
			&component.Model,
			&specsJSON,
			&component.Notes,
			&component.CreatedAt,
			&component.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan component: %w", err)
		}

		if err := json.Unmarshal([]byte(specsJSON), &component.Specs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal specs: %w", err)
		}

		components = append(components, &component)
	}

	return components, nil
}

func (r *componentRepo) Update(ctx context.Context, component *domain.Component) error {
	specsJSON, err := json.Marshal(component.Specs)
	if err != nil {
		return fmt.Errorf("failed to marshal specs: %w", err)
	}

	query := `
		UPDATE components
		SET name = ?, type = ?, manufacturer = ?, model = ?, specs = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		component.Name,
		component.Type,
		component.Manufacturer,
		component.Model,
		specsJSON,
		component.Notes,
		component.UpdatedAt,
		component.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("component not found")
	}

	return nil
}

func (r *componentRepo) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM components WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("component not found")
	}

	return nil
}
