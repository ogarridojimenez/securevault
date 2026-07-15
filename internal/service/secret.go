package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/repository"
)

type SecretService struct {
	secretRepo *repository.SecretRepository
	vaultRepo  *repository.VaultRepository
	masterKey  []byte
}

func NewSecretService(secretRepo *repository.SecretRepository, vaultRepo *repository.VaultRepository, masterKey []byte) *SecretService {
	return &SecretService{
		secretRepo: secretRepo,
		vaultRepo:  vaultRepo,
		masterKey:  masterKey,
	}
}

func (s *SecretService) Create(ctx context.Context, userID, vaultID, name, value string, metadata json.RawMessage) (*model.Secret, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: secret name is required", model.ErrInvalidInput)
	}
	if value == "" {
		return nil, fmt.Errorf("%w: secret value is required", model.ErrInvalidInput)
	}

	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}

	ciphertext, nonce, err := s.encrypt([]byte(value))
	if err != nil {
		return nil, fmt.Errorf("encrypt secret: %w", err)
	}

	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	return s.secretRepo.Create(ctx, vaultID, name, ciphertext, nonce, metadata)
}

func (s *SecretService) List(ctx context.Context, userID, vaultID string) (*model.SecretListResponse, error) {
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}

	secrets, err := s.secretRepo.ListByVault(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	if secrets == nil {
		secrets = []model.SecretSummary{}
	}
	return &model.SecretListResponse{Secrets: secrets, Total: len(secrets)}, nil
}

func (s *SecretService) Get(ctx context.Context, userID, vaultID, secretID string) (*model.SecretValueResponse, error) {
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}

	secret, err := s.secretRepo.FindByID(ctx, secretID)
	if err != nil {
		return nil, err
	}
	if secret.VaultID != vaultID {
		return nil, fmt.Errorf("%w: secret not found in this vault", model.ErrNotFound)
	}

	plaintext, err := s.decrypt(secret.Ciphertext, secret.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}

	return &model.SecretValueResponse{
		ID:    secret.ID,
		Name:  secret.Name,
		Value: string(plaintext),
	}, nil
}

func (s *SecretService) GetByName(ctx context.Context, userID, vaultID, name string) (*model.SecretValueResponse, error) {
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}

	secret, err := s.secretRepo.FindByName(ctx, vaultID, name)
	if err != nil {
		return nil, err
	}

	plaintext, err := s.decrypt(secret.Ciphertext, secret.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}

	return &model.SecretValueResponse{
		ID:    secret.ID,
		Name:  secret.Name,
		Value: string(plaintext),
	}, nil
}

func (s *SecretService) Delete(ctx context.Context, userID, vaultID, secretID string) error {
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}

	secret, err := s.secretRepo.FindByID(ctx, secretID)
	if err != nil {
		return err
	}
	if secret.VaultID != vaultID {
		return fmt.Errorf("%w: secret not found in this vault", model.ErrNotFound)
	}

	return s.secretRepo.Delete(ctx, secretID)
}

// AES-256-GCM encryption
func (s *SecretService) encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (s *SecretService) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
