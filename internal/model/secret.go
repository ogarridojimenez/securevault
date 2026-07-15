package model

import (
	"encoding/json"
	"time"
)

type Secret struct {
	ID        string          `json:"id"`
	VaultID   string          `json:"vault_id"`
	Name      string          `json:"name"`
	Value     string          `json:"value,omitempty"`
	Ciphertext []byte         `json:"-"`
	Nonce     []byte          `json:"-"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateSecretRequest struct {
	Name     string          `json:"name"`
	Value    string          `json:"value"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type UpdateSecretRequest struct {
	Value    string          `json:"value"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type SecretListResponse struct {
	Secrets []SecretSummary `json:"secrets"`
	Total   int             `json:"total"`
}

type SecretSummary struct {
	ID        string          `json:"id"`
	VaultID   string          `json:"vault_id"`
	Name      string          `json:"name"`
	Version   int             `json:"version"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type SecretValueResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
