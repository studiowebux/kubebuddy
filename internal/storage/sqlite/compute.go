package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type computeRepo struct {
	db *sql.DB
}

func (r *computeRepo) Create(ctx context.Context, compute *domain.Compute) error {
	tagsJSON, err := json.Marshal(compute.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	now := time.Now()
	compute.CreatedAt = now
	compute.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO computes (id, name, type, provider, region, tags, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, compute.ID, compute.Name, compute.Type, compute.Provider, compute.Region,
	   string(tagsJSON), compute.State, compute.CreatedAt, compute.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create compute: %w", err)
	}

	return nil
}

func (r *computeRepo) Get(ctx context.Context, id string) (*domain.Compute, error) {
	var compute domain.Compute
	var tagsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, type, provider, region, tags, state, created_at, updated_at
		FROM computes
		WHERE id = ?
	`, id).Scan(&compute.ID, &compute.Name, &compute.Type, &compute.Provider, &compute.Region,
		&tagsJSON, &compute.State, &compute.CreatedAt, &compute.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("compute not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get compute: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &compute.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &compute, nil
}

func (r *computeRepo) GetByNameProviderRegionType(ctx context.Context, name, provider, region, computeType string) (*domain.Compute, error) {
	var compute domain.Compute
	var tagsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, type, provider, region, tags, state, created_at, updated_at
		FROM computes
		WHERE name = ? AND provider = ? AND region = ? AND type = ?
	`, name, provider, region, computeType).Scan(&compute.ID, &compute.Name, &compute.Type, &compute.Provider, &compute.Region,
		&tagsJSON, &compute.State, &compute.CreatedAt, &compute.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get compute: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &compute.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &compute, nil
}

func (r *computeRepo) List(ctx context.Context, filters storage.ComputeFilters) ([]*domain.Compute, error) {
	query := `
		SELECT id, name, type, provider, region, tags, state, created_at, updated_at
		FROM computes
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if filters.Type != "" {
		query += " AND type = ?"
		args = append(args, filters.Type)
	}
	if filters.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filters.Provider)
	}
	if filters.Region != "" {
		query += " AND region = ?"
		args = append(args, filters.Region)
	}
	if filters.State != "" {
		query += " AND state = ?"
		args = append(args, filters.State)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list computes: %w", err)
	}
	defer rows.Close()

	computes := make([]*domain.Compute, 0)
	for rows.Next() {
		var compute domain.Compute
		var tagsJSON string

		err := rows.Scan(&compute.ID, &compute.Name, &compute.Type, &compute.Provider, &compute.Region,
			&tagsJSON, &compute.State, &compute.CreatedAt, &compute.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &compute.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		// Apply tag filters (post-query since tags are JSON)
		if len(filters.Tags) > 0 {
			match := true
			for key, value := range filters.Tags {
				if compute.Tags[key] != value {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		computes = append(computes, &compute)
	}

	return computes, nil
}

func (r *computeRepo) Update(ctx context.Context, compute *domain.Compute) error {
	tagsJSON, err := json.Marshal(compute.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	compute.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, `
		UPDATE computes
		SET name = ?, type = ?, provider = ?, region = ?, tags = ?, state = ?, updated_at = ?
		WHERE id = ?
	`, compute.Name, compute.Type, compute.Provider, compute.Region,
	   string(tagsJSON), compute.State, compute.UpdatedAt, compute.ID)

	if err != nil {
		return fmt.Errorf("failed to update compute: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("compute not found")
	}

	return nil
}

func (r *computeRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM computes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete compute: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("compute not found")
	}

	return nil
}
