package model

import "time"

type Vault struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateVaultRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateVaultRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type VaultListResponse struct {
	Vaults []Vault `json:"vaults"`
	Total  int     `json:"total"`
}
