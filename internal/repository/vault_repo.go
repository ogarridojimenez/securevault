package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ogarridojimenez/securevault/internal/model"
)

type VaultRepository struct {
	db *DB
}

func NewVaultRepository(db *DB) *VaultRepository {
	return &VaultRepository{db: db}
}

func (r *VaultRepository) Create(ctx context.Context, userID, name, description string) (*model.Vault, error) {
	v := &model.Vault{}
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO vaults (user_id, name, description) VALUES ($1, $2, $3)
		 RETURNING id, user_id, name, description, created_at, updated_at`,
		userID, name, description,
	).Scan(&v.ID, &v.UserID, &v.Name, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if pgxErrCode(err) == "23505" {
			return nil, fmt.Errorf("%w: vault name already exists", model.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("create vault: %w", err)
	}
	return v, nil
}

func (r *VaultRepository) ListByUser(ctx context.Context, userID string) ([]model.Vault, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at
		 FROM vaults WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list vaults: %w", err)
	}
	defer rows.Close()

	var vaults []model.Vault
	for rows.Next() {
		var v model.Vault
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.Description, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vault: %w", err)
		}
		vaults = append(vaults, v)
	}
	return vaults, rows.Err()
}

func (r *VaultRepository) FindByID(ctx context.Context, id string) (*model.Vault, error) {
	v := &model.Vault{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at FROM vaults WHERE id = $1`,
		id,
	).Scan(&v.ID, &v.UserID, &v.Name, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("find vault: %w", err)
	}
	return v, nil
}

func (r *VaultRepository) Update(ctx context.Context, id, name, description string) (*model.Vault, error) {
	v := &model.Vault{}
	err := r.db.Pool.QueryRow(ctx,
		`UPDATE vaults SET name = $1, description = $2, updated_at = now()
		 WHERE id = $3
		 RETURNING id, user_id, name, description, created_at, updated_at`,
		name, description, id,
	).Scan(&v.ID, &v.UserID, &v.Name, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update vault: %w", err)
	}
	return v, nil
}

func (r *VaultRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM vaults WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete vault: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *VaultRepository) BelongsToUser(ctx context.Context, vaultID, userID string) (bool, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM vaults WHERE id = $1 AND user_id = $2`,
		vaultID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check vault ownership: %w", err)
	}
	return count > 0, nil
}
