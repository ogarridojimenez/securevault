package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ogarridojimenez/securevault/internal/model"
)

type APIKeyRepository struct {
	db *DB
}

func NewAPIKeyRepository(db *DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, userID, name, keyPrefix, keyHash string) (*model.APIKey, error) {
	ak := &model.APIKey{}
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO api_keys (user_id, name, key_prefix, key_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, name, key_prefix, key_hash, last_used_at, revoked, created_at`,
		userID, name, keyPrefix, keyHash,
	).Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyPrefix, &ak.KeyHash, &ak.LastUsedAt, &ak.Revoked, &ak.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return ak, nil
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]model.APIKey, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, user_id, name, key_prefix, last_used_at, revoked, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var ak model.APIKey
		if err := rows.Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyPrefix, &ak.LastUsedAt, &ak.Revoked, &ak.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, ak)
	}
	return keys, rows.Err()
}

func (r *APIKeyRepository) FindByPrefix(ctx context.Context, keyPrefix string) (*model.APIKey, error) {
	ak := &model.APIKey{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, user_id, name, key_prefix, key_hash, last_used_at, revoked, created_at
		 FROM api_keys WHERE key_prefix = $1`,
		keyPrefix,
	).Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyPrefix, &ak.KeyHash, &ak.LastUsedAt, &ak.Revoked, &ak.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("find api key by prefix: %w", err)
	}
	return ak, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	tag, err := r.db.Pool.Exec(ctx,
		`UPDATE api_keys SET revoked = true WHERE id = $1 AND revoked = false`, id)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE api_keys SET last_used_at = now() WHERE id = $1`, id)
	return err
}
