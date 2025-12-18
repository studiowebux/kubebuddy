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

type ipAddressRepo struct {
	db *sql.DB
}

func (r *ipAddressRepo) Create(ctx context.Context, ip *domain.IPAddress) error {
	dnsJSON, err := json.Marshal(ip.DNSServers)
	if err != nil {
		return fmt.Errorf("failed to marshal dns_servers: %w", err)
	}

	query := `
		INSERT INTO ip_addresses (id, address, type, cidr, gateway, dns_servers, provider, region, vlan, notes, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		ip.ID,
		ip.Address,
		ip.Type,
		ip.CIDR,
		ip.Gateway,
		dnsJSON,
		ip.Provider,
		ip.Region,
		ip.VLAN,
		ip.Notes,
		ip.State,
		ip.CreatedAt,
		ip.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create IP address: %w", err)
	}

	return nil
}

func (r *ipAddressRepo) Get(ctx context.Context, id string) (*domain.IPAddress, error) {
	query := `
		SELECT id, address, type, cidr, gateway, dns_servers, provider, region, COALESCE(vlan, ''), notes, state, created_at, updated_at
		FROM ip_addresses
		WHERE id = ?
	`

	var ip domain.IPAddress
	var dnsJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ip.ID,
		&ip.Address,
		&ip.Type,
		&ip.CIDR,
		&ip.Gateway,
		&dnsJSON,
		&ip.Provider,
		&ip.Region,
		&ip.VLAN,
		&ip.Notes,
		&ip.State,
		&ip.CreatedAt,
		&ip.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("IP address not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP address: %w", err)
	}

	if err := json.Unmarshal([]byte(dnsJSON), &ip.DNSServers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dns_servers: %w", err)
	}

	return &ip, nil
}

func (r *ipAddressRepo) GetByAddress(ctx context.Context, address string) (*domain.IPAddress, error) {
	query := `
		SELECT id, address, type, cidr, gateway, dns_servers, provider, region, COALESCE(vlan, ''), notes, state, created_at, updated_at
		FROM ip_addresses
		WHERE address = ?
	`

	var ip domain.IPAddress
	var dnsJSON string

	err := r.db.QueryRowContext(ctx, query, address).Scan(
		&ip.ID,
		&ip.Address,
		&ip.Type,
		&ip.CIDR,
		&ip.Gateway,
		&dnsJSON,
		&ip.Provider,
		&ip.Region,
		&ip.VLAN,
		&ip.Notes,
		&ip.State,
		&ip.CreatedAt,
		&ip.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP address: %w", err)
	}

	if err := json.Unmarshal([]byte(dnsJSON), &ip.DNSServers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dns_servers: %w", err)
	}

	return &ip, nil
}

func (r *ipAddressRepo) List(ctx context.Context, filters storage.IPAddressFilters) ([]*domain.IPAddress, error) {
	query := "SELECT id, address, type, cidr, gateway, dns_servers, provider, region, COALESCE(vlan, ''), notes, state, created_at, updated_at FROM ip_addresses WHERE 1=1"
	args := []interface{}{}

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

	query += " ORDER BY address"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list IP addresses: %w", err)
	}
	defer rows.Close()

	var ips []*domain.IPAddress
	for rows.Next() {
		var ip domain.IPAddress
		var dnsJSON string

		err := rows.Scan(
			&ip.ID,
			&ip.Address,
			&ip.Type,
			&ip.CIDR,
			&ip.Gateway,
			&dnsJSON,
			&ip.Provider,
			&ip.Region,
			&ip.VLAN,
			&ip.Notes,
			&ip.State,
			&ip.CreatedAt,
			&ip.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP address: %w", err)
		}

		if err := json.Unmarshal([]byte(dnsJSON), &ip.DNSServers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dns_servers: %w", err)
		}

		ips = append(ips, &ip)
	}

	return ips, nil
}

func (r *ipAddressRepo) Update(ctx context.Context, ip *domain.IPAddress) error {
	dnsJSON, err := json.Marshal(ip.DNSServers)
	if err != nil {
		return fmt.Errorf("failed to marshal dns_servers: %w", err)
	}

	query := `
		UPDATE ip_addresses
		SET address = ?, type = ?, cidr = ?, gateway = ?, dns_servers = ?, provider = ?, region = ?, vlan = ?, notes = ?, state = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		ip.Address,
		ip.Type,
		ip.CIDR,
		ip.Gateway,
		dnsJSON,
		ip.Provider,
		ip.Region,
		ip.VLAN,
		ip.Notes,
		ip.State,
		ip.UpdatedAt,
		ip.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update IP address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP address not found")
	}

	return nil
}

func (r *ipAddressRepo) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM ip_addresses WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete IP address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP address not found")
	}

	return nil
}

type computeIPRepo struct {
	db *sql.DB
}

func (r *computeIPRepo) Assign(ctx context.Context, assignment *domain.ComputeIP) error {
	query := `
		INSERT INTO compute_ips (id, compute_id, ip_id, interface_name, is_primary, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	isPrimary := 0
	if assignment.IsPrimary {
		isPrimary = 1
	}

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.ComputeID,
		assignment.IPID,
		assignment.InterfaceName,
		isPrimary,
		assignment.CreatedAt,
		assignment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to assign IP to compute: %w", err)
	}

	return nil
}

func (r *computeIPRepo) Unassign(ctx context.Context, id string) error {
	query := "DELETE FROM compute_ips WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unassign IP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP assignment not found")
	}

	return nil
}

func (r *computeIPRepo) UnassignByIP(ctx context.Context, ipID string) error {
	query := "DELETE FROM compute_ips WHERE ip_id = ?"

	_, err := r.db.ExecContext(ctx, query, ipID)
	if err != nil {
		return fmt.Errorf("failed to unassign IP: %w", err)
	}

	return nil
}

func (r *computeIPRepo) ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeIP, error) {
	query := `
		SELECT id, compute_id, ip_id, COALESCE(interface_name, ''), is_primary, created_at, updated_at
		FROM compute_ips
		WHERE compute_id = ?
		ORDER BY is_primary DESC, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, computeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute IPs: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeIP
	for rows.Next() {
		var assignment domain.ComputeIP
		var isPrimary int

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.IPID,
			&assignment.InterfaceName,
			&isPrimary,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute IP: %w", err)
		}

		assignment.IsPrimary = isPrimary == 1

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeIPRepo) ListByIP(ctx context.Context, ipID string) ([]*domain.ComputeIP, error) {
	query := `
		SELECT id, compute_id, ip_id, COALESCE(interface_name, ''), is_primary, created_at, updated_at
		FROM compute_ips
		WHERE ip_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, ipID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute IPs: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeIP
	for rows.Next() {
		var assignment domain.ComputeIP
		var isPrimary int

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.IPID,
			&assignment.InterfaceName,
			&isPrimary,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute IP: %w", err)
		}

		assignment.IsPrimary = isPrimary == 1

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeIPRepo) List(ctx context.Context) ([]*domain.ComputeIP, error) {
	query := `
		SELECT id, compute_id, ip_id, COALESCE(interface_name, ''), is_primary, created_at, updated_at
		FROM compute_ips
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all compute IPs: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.ComputeIP
	for rows.Next() {
		var assignment domain.ComputeIP
		var isPrimary int

		err := rows.Scan(
			&assignment.ID,
			&assignment.ComputeID,
			&assignment.IPID,
			&assignment.InterfaceName,
			&isPrimary,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compute IP: %w", err)
		}

		assignment.IsPrimary = isPrimary == 1

		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

func (r *computeIPRepo) GetPrimaryIP(ctx context.Context, computeID string) (*domain.ComputeIP, error) {
	query := `
		SELECT id, compute_id, ip_id, is_primary, created_at, updated_at
		FROM compute_ips
		WHERE compute_id = ? AND is_primary = 1
		LIMIT 1
	`

	var assignment domain.ComputeIP
	var isPrimary int

	err := r.db.QueryRowContext(ctx, query, computeID).Scan(
		&assignment.ID,
		&assignment.ComputeID,
		&assignment.IPID,
		&isPrimary,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No primary IP found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get primary IP: %w", err)
	}

	assignment.IsPrimary = isPrimary == 1

	return &assignment, nil
}

func (r *computeIPRepo) GetByComputeAndIP(ctx context.Context, computeID, ipID string) (*domain.ComputeIP, error) {
	query := `
		SELECT id, compute_id, ip_id, is_primary, created_at, updated_at
		FROM compute_ips
		WHERE compute_id = ? AND ip_id = ?
	`

	var assignment domain.ComputeIP
	var isPrimary int

	err := r.db.QueryRowContext(ctx, query, computeID, ipID).Scan(
		&assignment.ID,
		&assignment.ComputeID,
		&assignment.IPID,
		&isPrimary,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get compute IP: %w", err)
	}

	assignment.IsPrimary = isPrimary == 1

	return &assignment, nil
}

func (r *computeIPRepo) UpdatePrimary(ctx context.Context, id string, isPrimary bool) error {
	query := "UPDATE compute_ips SET is_primary = ?, updated_at = ? WHERE id = ?"

	isPrimaryInt := 0
	if isPrimary {
		isPrimaryInt = 1
	}

	result, err := r.db.ExecContext(ctx, query, isPrimaryInt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update primary flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP assignment not found")
	}

	return nil
}
