package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type dnsRecordRepo struct {
	db *sql.DB
}

func (r *dnsRecordRepo) Create(ctx context.Context, record *domain.DNSRecord) error {
	query := `
		INSERT INTO dns_records (id, name, type, value, ip_id, ttl, zone, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var ipID interface{}
	if record.IPID != "" {
		ipID = record.IPID
	}

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.Name,
		record.Type,
		record.Value,
		ipID,
		record.TTL,
		record.Zone,
		record.Notes,
		record.CreatedAt,
		record.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	return nil
}

func (r *dnsRecordRepo) Get(ctx context.Context, id string) (*domain.DNSRecord, error) {
	query := `
		SELECT id, name, type, value, ip_id, ttl, zone, notes, created_at, updated_at
		FROM dns_records
		WHERE id = ?
	`

	var record domain.DNSRecord
	var ipID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.Name,
		&record.Type,
		&record.Value,
		&ipID,
		&record.TTL,
		&record.Zone,
		&record.Notes,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("DNS record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	if ipID.Valid {
		record.IPID = ipID.String
	}

	return &record, nil
}

func (r *dnsRecordRepo) GetByNameTypeZone(ctx context.Context, name, recordType, zone string) (*domain.DNSRecord, error) {
	query := `
		SELECT id, name, type, value, ip_id, ttl, zone, notes, created_at, updated_at
		FROM dns_records
		WHERE name = ? AND type = ? AND zone = ?
	`

	var record domain.DNSRecord
	var ipID sql.NullString

	err := r.db.QueryRowContext(ctx, query, name, recordType, zone).Scan(
		&record.ID,
		&record.Name,
		&record.Type,
		&record.Value,
		&ipID,
		&record.TTL,
		&record.Zone,
		&record.Notes,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for upsert logic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	if ipID.Valid {
		record.IPID = ipID.String
	}

	return &record, nil
}

func (r *dnsRecordRepo) List(ctx context.Context, filters storage.DNSRecordFilters) ([]*domain.DNSRecord, error) {
	query := "SELECT id, name, type, value, ip_id, ttl, zone, notes, created_at, updated_at FROM dns_records WHERE 1=1"
	args := []interface{}{}

	if filters.Type != "" {
		query += " AND type = ?"
		args = append(args, filters.Type)
	}

	if filters.Zone != "" {
		query += " AND zone = ?"
		args = append(args, filters.Zone)
	}

	if filters.IPID != "" {
		query += " AND ip_id = ?"
		args = append(args, filters.IPID)
	}

	if filters.Name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filters.Name+"%")
	}

	query += " ORDER BY zone, name, type"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}
	defer rows.Close()

	var records []*domain.DNSRecord
	for rows.Next() {
		var record domain.DNSRecord
		var ipID sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.Type,
			&record.Value,
			&ipID,
			&record.TTL,
			&record.Zone,
			&record.Notes,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DNS record: %w", err)
		}

		if ipID.Valid {
			record.IPID = ipID.String
		}

		records = append(records, &record)
	}

	return records, nil
}

func (r *dnsRecordRepo) Update(ctx context.Context, record *domain.DNSRecord) error {
	query := `
		UPDATE dns_records
		SET name = ?, type = ?, value = ?, ip_id = ?, ttl = ?, zone = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	var ipID interface{}
	if record.IPID != "" {
		ipID = record.IPID
	}

	result, err := r.db.ExecContext(ctx, query,
		record.Name,
		record.Type,
		record.Value,
		ipID,
		record.TTL,
		record.Zone,
		record.Notes,
		record.UpdatedAt,
		record.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DNS record not found")
	}

	return nil
}

func (r *dnsRecordRepo) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM dns_records WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DNS record not found")
	}

	return nil
}
