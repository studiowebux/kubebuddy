package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type portAssignmentRepo struct {
	db *sql.DB
}

func (r *portAssignmentRepo) Create(ctx context.Context, assignment *domain.PortAssignment) error {
	query := `
		INSERT INTO port_assignments (id, assignment_id, ip_id, port, protocol, service_port, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.AssignmentID,
		assignment.IPID,
		assignment.Port,
		assignment.Protocol,
		assignment.ServicePort,
		assignment.Description,
		assignment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create port assignment: %w", err)
	}

	return nil
}

func (r *portAssignmentRepo) Get(ctx context.Context, id string) (*domain.PortAssignment, error) {
	query := `
		SELECT id, assignment_id, ip_id, port, protocol, service_port, description, created_at
		FROM port_assignments
		WHERE id = ?
	`

	var assignment domain.PortAssignment

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID,
		&assignment.AssignmentID,
		&assignment.IPID,
		&assignment.Port,
		&assignment.Protocol,
		&assignment.ServicePort,
		&assignment.Description,
		&assignment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("port assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get port assignment: %w", err)
	}

	return &assignment, nil
}

func (r *portAssignmentRepo) GetByIPPortProtocol(ctx context.Context, ipID string, port int, protocol string) (*domain.PortAssignment, error) {
	query := `
		SELECT id, assignment_id, ip_id, port, protocol, service_port, description, created_at
		FROM port_assignments
		WHERE ip_id = ? AND port = ? AND protocol = ?
	`

	var assignment domain.PortAssignment

	err := r.db.QueryRowContext(ctx, query, ipID, port, protocol).Scan(
		&assignment.ID,
		&assignment.AssignmentID,
		&assignment.IPID,
		&assignment.Port,
		&assignment.Protocol,
		&assignment.ServicePort,
		&assignment.Description,
		&assignment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get port assignment: %w", err)
	}

	return &assignment, nil
}

func (r *portAssignmentRepo) List(ctx context.Context, filters storage.PortAssignmentFilters) ([]*domain.PortAssignment, error) {
	query := "SELECT id, assignment_id, ip_id, port, protocol, service_port, description, created_at FROM port_assignments WHERE 1=1"
	args := []interface{}{}

	if filters.AssignmentID != "" {
		query += " AND assignment_id = ?"
		args = append(args, filters.AssignmentID)
	}

	if filters.IPID != "" {
		query += " AND ip_id = ?"
		args = append(args, filters.IPID)
	}

	if filters.Protocol != "" {
		query += " AND protocol = ?"
		args = append(args, filters.Protocol)
	}

	query += " ORDER BY ip_id, port"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list port assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.PortAssignment
	for rows.Next() {
		var assignment domain.PortAssignment

		err := rows.Scan(
			&assignment.ID,
			&assignment.AssignmentID,
			&assignment.IPID,
			&assignment.Port,
			&assignment.Protocol,
			&assignment.ServicePort,
			&assignment.Description,
			&assignment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan port assignment: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *portAssignmentRepo) Update(ctx context.Context, assignment *domain.PortAssignment) error {
	query := `
		UPDATE port_assignments
		SET assignment_id = ?, ip_id = ?, port = ?, protocol = ?, service_port = ?, description = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		assignment.AssignmentID,
		assignment.IPID,
		assignment.Port,
		assignment.Protocol,
		assignment.ServicePort,
		assignment.Description,
		assignment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update port assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("port assignment not found")
	}

	return nil
}

func (r *portAssignmentRepo) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM port_assignments WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete port assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("port assignment not found")
	}

	return nil
}

func (r *portAssignmentRepo) DeleteByAssignment(ctx context.Context, assignmentID string) error {
	query := "DELETE FROM port_assignments WHERE assignment_id = ?"

	_, err := r.db.ExecContext(ctx, query, assignmentID)
	if err != nil {
		return fmt.Errorf("failed to delete port assignments: %w", err)
	}

	return nil
}
