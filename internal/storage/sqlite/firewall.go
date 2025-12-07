package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type firewallRuleRepo struct {
	db *sql.DB
}

func (r *firewallRuleRepo) Create(ctx context.Context, rule *domain.FirewallRule) error {
	query := `
		INSERT INTO firewall_rules (id, name, action, protocol, source, destination, port_start, port_end, description, priority, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var portStart, portEnd interface{}
	if rule.PortStart != nil {
		portStart = *rule.PortStart
	}
	if rule.PortEnd != nil {
		portEnd = *rule.PortEnd
	}

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.Action,
		rule.Protocol,
		rule.Source,
		rule.Destination,
		portStart,
		portEnd,
		rule.Description,
		rule.Priority,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create firewall rule: %w", err)
	}

	return nil
}

func (r *firewallRuleRepo) Get(ctx context.Context, id string) (*domain.FirewallRule, error) {
	query := `
		SELECT id, name, action, protocol, source, destination, port_start, port_end, description, priority, created_at, updated_at
		FROM firewall_rules
		WHERE id = ?
	`

	var rule domain.FirewallRule
	var portStart, portEnd sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Action,
		&rule.Protocol,
		&rule.Source,
		&rule.Destination,
		&portStart,
		&portEnd,
		&rule.Description,
		&rule.Priority,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("firewall rule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall rule: %w", err)
	}

	if portStart.Valid {
		ps := int(portStart.Int64)
		rule.PortStart = &ps
	}
	if portEnd.Valid {
		pe := int(portEnd.Int64)
		rule.PortEnd = &pe
	}

	return &rule, nil
}

func (r *firewallRuleRepo) GetByName(ctx context.Context, name string) (*domain.FirewallRule, error) {
	query := `
		SELECT id, name, action, protocol, source, destination, port_start, port_end, description, priority, created_at, updated_at
		FROM firewall_rules
		WHERE name = ?
	`

	var rule domain.FirewallRule
	var portStart, portEnd sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Action,
		&rule.Protocol,
		&rule.Source,
		&rule.Destination,
		&portStart,
		&portEnd,
		&rule.Description,
		&rule.Priority,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall rule: %w", err)
	}

	if portStart.Valid {
		ps := int(portStart.Int64)
		rule.PortStart = &ps
	}
	if portEnd.Valid {
		pe := int(portEnd.Int64)
		rule.PortEnd = &pe
	}

	return &rule, nil
}

func (r *firewallRuleRepo) List(ctx context.Context, filters storage.FirewallRuleFilters) ([]*domain.FirewallRule, error) {
	query := "SELECT id, name, action, protocol, source, destination, port_start, port_end, description, priority, created_at, updated_at FROM firewall_rules WHERE 1=1"
	args := []interface{}{}

	if filters.Action != "" {
		query += " AND action = ?"
		args = append(args, filters.Action)
	}

	if filters.Protocol != "" {
		query += " AND protocol = ?"
		args = append(args, filters.Protocol)
	}

	query += " ORDER BY priority, name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list firewall rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.FirewallRule
	for rows.Next() {
		var rule domain.FirewallRule
		var portStart, portEnd sql.NullInt64

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Action,
			&rule.Protocol,
			&rule.Source,
			&rule.Destination,
			&portStart,
			&portEnd,
			&rule.Description,
			&rule.Priority,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan firewall rule: %w", err)
		}

		if portStart.Valid {
			ps := int(portStart.Int64)
			rule.PortStart = &ps
		}
		if portEnd.Valid {
			pe := int(portEnd.Int64)
			rule.PortEnd = &pe
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

func (r *firewallRuleRepo) Update(ctx context.Context, rule *domain.FirewallRule) error {
	query := `
		UPDATE firewall_rules
		SET name = ?, action = ?, protocol = ?, source = ?, destination = ?, port_start = ?, port_end = ?, description = ?, priority = ?, updated_at = ?
		WHERE id = ?
	`

	var portStart, portEnd interface{}
	if rule.PortStart != nil {
		portStart = *rule.PortStart
	}
	if rule.PortEnd != nil {
		portEnd = *rule.PortEnd
	}

	result, err := r.db.ExecContext(ctx, query,
		rule.Name,
		rule.Action,
		rule.Protocol,
		rule.Source,
		rule.Destination,
		portStart,
		portEnd,
		rule.Description,
		rule.Priority,
		rule.UpdatedAt,
		rule.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update firewall rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("firewall rule not found")
	}

	return nil
}

func (r *firewallRuleRepo) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM firewall_rules WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete firewall rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("firewall rule not found")
	}

	return nil
}

type computeFirewallRuleRepo struct {
	db *sql.DB
}

func (r *computeFirewallRuleRepo) Assign(ctx context.Context, assignment *domain.ComputeFirewallRule) error {
	query := `
		INSERT INTO compute_firewall_rules (id, compute_id, rule_id, enabled, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.ComputeID,
		assignment.RuleID,
		assignment.Enabled,
		assignment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to assign firewall rule: %w", err)
	}

	return nil
}

func (r *computeFirewallRuleRepo) Unassign(ctx context.Context, id string) error {
	query := "DELETE FROM compute_firewall_rules WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unassign firewall rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("firewall rule assignment not found")
	}

	return nil
}

func (r *computeFirewallRuleRepo) ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeFirewallRule, error) {
	query := `
		SELECT id, compute_id, rule_id, enabled, created_at
		FROM compute_firewall_rules
		WHERE compute_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, computeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute firewall rules: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeFirewallRule
	for rows.Next() {
		var assignment domain.ComputeFirewallRule

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.RuleID,
			&assignment.Enabled,
			&assignment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute firewall rule: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeFirewallRuleRepo) ListByRule(ctx context.Context, ruleID string) ([]*domain.ComputeFirewallRule, error) {
	query := `
		SELECT id, compute_id, rule_id, enabled, created_at
		FROM compute_firewall_rules
		WHERE rule_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute firewall rules: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeFirewallRule
	for rows.Next() {
		var assignment domain.ComputeFirewallRule

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.RuleID,
			&assignment.Enabled,
			&assignment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute firewall rule: %w", err)
		}

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeFirewallRuleRepo) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	query := "UPDATE compute_firewall_rules SET enabled = ? WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, enabled, id)
	if err != nil {
		return fmt.Errorf("failed to update firewall rule enabled status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("firewall rule assignment not found")
	}

	return nil
}
