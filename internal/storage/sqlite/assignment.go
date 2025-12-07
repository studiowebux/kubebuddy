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

type assignmentRepo struct {
	db *sql.DB
}

func (r *assignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	allocatedJSON, err := json.Marshal(assignment.Allocated)
	if err != nil {
		return fmt.Errorf("failed to marshal allocated: %w", err)
	}

	now := time.Now()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO assignments (id, service_id, compute_id, allocated, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, assignment.ID, assignment.ServiceID, assignment.ComputeID,
	   string(allocatedJSON), assignment.CreatedAt, assignment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

func (r *assignmentRepo) Get(ctx context.Context, id string) (*domain.Assignment, error) {
	var assignment domain.Assignment
	var allocatedJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, service_id, compute_id, allocated, created_at, updated_at
		FROM assignments
		WHERE id = ?
	`, id).Scan(&assignment.ID, &assignment.ServiceID, &assignment.ComputeID,
		&allocatedJSON, &assignment.CreatedAt, &assignment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	if err := json.Unmarshal([]byte(allocatedJSON), &assignment.Allocated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allocated: %w", err)
	}

	return &assignment, nil
}

func (r *assignmentRepo) GetByComputeAndService(ctx context.Context, computeID, serviceID string) (*domain.Assignment, error) {
	var assignment domain.Assignment
	var allocatedJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, service_id, compute_id, allocated, created_at, updated_at
		FROM assignments
		WHERE compute_id = ? AND service_id = ?
	`, computeID, serviceID).Scan(&assignment.ID, &assignment.ServiceID, &assignment.ComputeID,
		&allocatedJSON, &assignment.CreatedAt, &assignment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil for upsert logic
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	if err := json.Unmarshal([]byte(allocatedJSON), &assignment.Allocated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allocated: %w", err)
	}

	return &assignment, nil
}

func (r *assignmentRepo) List(ctx context.Context, filters storage.AssignmentFilters) ([]*domain.Assignment, error) {
	query := `
		SELECT id, service_id, compute_id, allocated, created_at, updated_at
		FROM assignments
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if filters.ServiceID != "" {
		query += " AND service_id = ?"
		args = append(args, filters.ServiceID)
	}
	if filters.ComputeID != "" {
		query += " AND compute_id = ?"
		args = append(args, filters.ComputeID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]*domain.Assignment, 0)
	for rows.Next() {
		var assignment domain.Assignment
		var allocatedJSON string

		err := rows.Scan(&assignment.ID, &assignment.ServiceID, &assignment.ComputeID,
			&allocatedJSON, &assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		if err := json.Unmarshal([]byte(allocatedJSON), &assignment.Allocated); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allocated: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *assignmentRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	allocatedJSON, err := json.Marshal(assignment.Allocated)
	if err != nil {
		return fmt.Errorf("failed to marshal allocated: %w", err)
	}

	assignment.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, `
		UPDATE assignments
		SET allocated = ?, updated_at = ?
		WHERE id = ?
	`, string(allocatedJSON), assignment.UpdatedAt, assignment.ID)

	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	return nil
}

func (r *assignmentRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM assignments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("assignment not found")
	}

	return nil
}

func (r *assignmentRepo) DeleteByService(ctx context.Context, serviceID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM assignments WHERE service_id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete assignments by service: %w", err)
	}

	return nil
}

func (r *assignmentRepo) DeleteByCompute(ctx context.Context, computeID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM assignments WHERE compute_id = ?", computeID)
	if err != nil {
		return fmt.Errorf("failed to delete assignments by compute: %w", err)
	}

	return nil
}
