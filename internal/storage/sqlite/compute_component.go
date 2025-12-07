package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/studiowebux/kubebuddy/internal/domain"
)

type computeComponentRepo struct {
	db *sql.DB
}

func (r *computeComponentRepo) Assign(ctx context.Context, assignment *domain.ComputeComponent) error {
	query := `
		INSERT INTO compute_components (id, compute_id, component_id, quantity, slot, serial_no, notes, raid_level, raid_group, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.ComputeID,
		assignment.ComponentID,
		assignment.Quantity,
		assignment.Slot,
		assignment.SerialNo,
		assignment.Notes,
		assignment.RaidLevel,
		assignment.RaidGroup,
		assignment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to assign component: %w", err)
	}

	return nil
}

func (r *computeComponentRepo) Unassign(ctx context.Context, id string) error {
	query := "DELETE FROM compute_components WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unassign component: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("assignment not found")
	}

	return nil
}

func (r *computeComponentRepo) ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeComponent, error) {
	query := `
		SELECT id, compute_id, component_id, quantity, slot, serial_no, notes, raid_level, raid_group, created_at
		FROM compute_components
		WHERE compute_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, computeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute components: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeComponent
	for rows.Next() {
		var assignment domain.ComputeComponent

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.ComponentID,
			&assignment.Quantity,
			&assignment.Slot,
			&assignment.SerialNo,
			&assignment.Notes,
			&assignment.RaidLevel,
			&assignment.RaidGroup,
			&assignment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute component: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeComponentRepo) ListByComponent(ctx context.Context, componentID string) ([]*domain.ComputeComponent, error) {
	query := `
		SELECT id, compute_id, component_id, quantity, slot, serial_no, notes, raid_level, raid_group, created_at
		FROM compute_components
		WHERE component_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list component assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeComponent
	for rows.Next() {
		var assignment domain.ComputeComponent

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.ComponentID,
			&assignment.Quantity,
			&assignment.Slot,
			&assignment.SerialNo,
			&assignment.Notes,
			&assignment.RaidLevel,
			&assignment.RaidGroup,
			&assignment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan component assignment: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}
