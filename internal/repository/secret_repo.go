package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ogarridojimenez/securevault/internal/model"
)

type SecretRepository struct {
	db *DB
}

func NewSecretRepository(db *DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) Create(ctx context.Context, vaultID, name string, ciphertext, nonce []byte, metadata []byte) (*model.Secret, error) {
	s := &model.Secret{Metadata: metadata}
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO secrets (vault_id, name, ciphertext, nonce, metadata)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, vault_id, name, ciphertext, nonce, metadata, version, created_at, updated_at`,
		vaultID, name, ciphertext, nonce, metadata,
	).Scan(&s.ID, &s.VaultID, &s.Name, &s.Ciphertext, &s.Nonce, &s.Metadata, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if pgxErrCode(err) == "23505" {
			return nil, fmt.Errorf("%w: secret name already exists in vault", model.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("create secret: %w", err)
	}
	return s, nil
}

func (r *SecretRepository) ListByVault(ctx context.Context, vaultID string) ([]model.SecretSummary, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, vault_id, name, metadata, version, created_at, updated_at
		 FROM secrets WHERE vault_id = $1 ORDER BY name`,
		vaultID,
	)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	var secrets []model.SecretSummary
	for rows.Next() {
		var s model.SecretSummary
		if err := rows.Scan(&s.ID, &s.VaultID, &s.Name, &s.Metadata, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}

func (r *SecretRepository) FindByID(ctx context.Context, id string) (*model.Secret, error) {
	s := &model.Secret{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, vault_id, name, ciphertext, nonce, metadata, version, created_at, updated_at
		 FROM secrets WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.VaultID, &s.Name, &s.Ciphertext, &s.Nonce, &s.Metadata, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("find secret: %w", err)
	}
	return s, nil
}

func (r *SecretRepository) FindByName(ctx context.Context, vaultID, name string) (*model.Secret, error) {
	s := &model.Secret{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, vault_id, name, ciphertext, nonce, metadata, version, created_at, updated_at
		 FROM secrets WHERE vault_id = $1 AND name = $2`,
		vaultID, name,
	).Scan(&s.ID, &s.VaultID, &s.Name, &s.Ciphertext, &s.Nonce, &s.Metadata, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("find secret by name: %w", err)
	}
	return s, nil
}

func (r *SecretRepository) Update(ctx context.Context, id string, ciphertext, nonce []byte, metadata []byte) (*model.Secret, error) {
	s := &model.Secret{}
	err := r.db.Pool.QueryRow(ctx,
		`UPDATE secrets SET ciphertext = $1, nonce = $2, metadata = $3, version = version + 1, updated_at = now()
		 WHERE id = $4
		 RETURNING id, vault_id, name, ciphertext, nonce, metadata, version, created_at, updated_at`,
		ciphertext, nonce, metadata, id,
	).Scan(&s.ID, &s.VaultID, &s.Name, &s.Ciphertext, &s.Nonce, &s.Metadata, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update secret: %w", err)
	}
	return s, nil
}

func (r *SecretRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM secrets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *SecretRepository) CountByVault(ctx context.Context, vaultID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM secrets WHERE vault_id = $1`, vaultID,
	).Scan(&count)
	return count, err
}
