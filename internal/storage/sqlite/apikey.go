package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
)

type apikeyRepo struct {
	db *sql.DB
}

func (r *apikeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	key.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO api_keys (id, name, key_hash, scope, description, created_by, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, key.ID, key.Name, key.KeyHash, key.Scope, key.Description, key.CreatedBy, key.CreatedAt, key.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

func (r *apikeyRepo) Get(ctx context.Context, id string) (*domain.APIKey, error) {
	var key domain.APIKey
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, key_hash, scope, description, created_by, created_at, expires_at
		FROM api_keys
		WHERE id = ?
	`, id).Scan(&key.ID, &key.Name, &key.KeyHash, &key.Scope, &key.Description,
		&key.CreatedBy, &key.CreatedAt, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return &key, nil
}

func (r *apikeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var key domain.APIKey
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, key_hash, scope, description, created_by, created_at, expires_at
		FROM api_keys
		WHERE key_hash = ?
	`, keyHash).Scan(&key.ID, &key.Name, &key.KeyHash, &key.Scope, &key.Description,
		&key.CreatedBy, &key.CreatedAt, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return &key, nil
}

func (r *apikeyRepo) List(ctx context.Context) ([]*domain.APIKey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, key_hash, scope, description, created_by, created_at, expires_at
		FROM api_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	keys := make([]*domain.APIKey, 0)
	for rows.Next() {
		var key domain.APIKey
		var expiresAt sql.NullTime

		err := rows.Scan(&key.ID, &key.Name, &key.KeyHash, &key.Scope, &key.Description,
			&key.CreatedBy, &key.CreatedAt, &expiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}

		keys = append(keys, &key)
	}

	return keys, nil
}

func (r *apikeyRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}
