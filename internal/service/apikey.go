package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/repository"
)

type APIKeyService struct {
	apiKeyRepo *repository.APIKeyRepository
}

func NewAPIKeyService(apiKeyRepo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{apiKeyRepo: apiKeyRepo}
}

func (s *APIKeyService) Create(ctx context.Context, userID, name string) (*model.CreateAPIKeyResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: api key name is required", model.ErrInvalidInput)
	}

	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	fullKey := "sv_" + hex.EncodeToString(keyBytes)

	keyHash := sha256Hex([]byte(fullKey))
	prefix := fullKey[:10]

	aKey, err := s.apiKeyRepo.Create(ctx, userID, name, prefix, keyHash)
	if err != nil {
		return nil, err
	}

	resp := &model.CreateAPIKeyResponse{}
	resp.APIKey = *aKey
	resp.APIKey.FullKey = fullKey
	resp.FullKey = fullKey
	return resp, nil
}

func (s *APIKeyService) List(ctx context.Context, userID string) ([]model.APIKey, error) {
	keys, err := s.apiKeyRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if keys == nil {
		keys = []model.APIKey{}
	}
	return keys, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, userID, keyID string) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}

func (s *APIKeyService) Validate(ctx context.Context, apiKey string) (string, error) {
	if len(apiKey) < 10 || apiKey[:3] != "sv_" {
		return "", fmt.Errorf("%w: invalid api key format", model.ErrUnauthorized)
	}

	prefix := apiKey[:10]
	aKey, err := s.apiKeyRepo.FindByPrefix(ctx, prefix)
	if err != nil {
		return "", fmt.Errorf("%w: invalid api key", model.ErrUnauthorized)
	}

	if aKey.Revoked {
		return "", fmt.Errorf("%w: api key revoked", model.ErrUnauthorized)
	}

	keyHash := sha256Hex([]byte(apiKey))
	if keyHash != aKey.KeyHash {
		return "", fmt.Errorf("%w: invalid api key", model.ErrUnauthorized)
	}

	go func() {
		_ = s.apiKeyRepo.UpdateLastUsed(context.Background(), aKey.ID)
	}()

	return aKey.UserID, nil
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
