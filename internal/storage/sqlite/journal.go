package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type journalRepo struct {
	db *sql.DB
}

func (r *journalRepo) Create(ctx context.Context, entry *domain.JournalEntry) error {
	entry.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO journal_entries (id, compute_id, category, content, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.ComputeID, entry.Category, entry.Content, entry.CreatedBy, entry.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	return nil
}

func (r *journalRepo) Get(ctx context.Context, id string) (*domain.JournalEntry, error) {
	var entry domain.JournalEntry

	err := r.db.QueryRowContext(ctx, `
		SELECT id, compute_id, category, content, created_by, created_at
		FROM journal_entries
		WHERE id = ?
	`, id).Scan(&entry.ID, &entry.ComputeID, &entry.Category, &entry.Content, &entry.CreatedBy, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("journal entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get journal entry: %w", err)
	}

	return &entry, nil
}

func (r *journalRepo) List(ctx context.Context, filters storage.JournalFilters) ([]*domain.JournalEntry, error) {
	query := `
		SELECT id, compute_id, category, content, created_by, created_at
		FROM journal_entries
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if filters.ComputeID != "" {
		query += " AND compute_id = ?"
		args = append(args, filters.ComputeID)
	}
	if filters.Category != "" {
		query += " AND category = ?"
		args = append(args, filters.Category)
	}
	if filters.From != nil {
		query += " AND created_at >= ?"
		args = append(args, filters.From)
	}
	if filters.To != nil {
		query += " AND created_at <= ?"
		args = append(args, filters.To)
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list journal entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*domain.JournalEntry, 0)
	for rows.Next() {
		var entry domain.JournalEntry

		err := rows.Scan(&entry.ID, &entry.ComputeID, &entry.Category, &entry.Content, &entry.CreatedBy, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal entry: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *journalRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM journal_entries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete journal entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("journal entry not found")
	}

	return nil
}
