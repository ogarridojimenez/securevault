package service

import (
	"context"
	"fmt"

	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/repository"
)

type VaultService struct {
	vaultRepo *repository.VaultRepository
}

func NewVaultService(vaultRepo *repository.VaultRepository) *VaultService {
	return &VaultService{vaultRepo: vaultRepo}
}

func (s *VaultService) Create(ctx context.Context, userID, name, description string) (*model.Vault, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: vault name is required", model.ErrInvalidInput)
	}
	return s.vaultRepo.Create(ctx, userID, name, description)
}

func (s *VaultService) List(ctx context.Context, userID string) (*model.VaultListResponse, error) {
	vaults, err := s.vaultRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if vaults == nil {
		vaults = []model.Vault{}
	}
	return &model.VaultListResponse{Vaults: vaults, Total: len(vaults)}, nil
}

func (s *VaultService) Get(ctx context.Context, userID, vaultID string) (*model.Vault, error) {
	vault, err := s.vaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	if vault.UserID != userID {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}
	return vault, nil
}

func (s *VaultService) Update(ctx context.Context, userID, vaultID, name, description string) (*model.Vault, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: vault name is required", model.ErrInvalidInput)
	}
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}
	return s.vaultRepo.Update(ctx, vaultID, name, description)
}

func (s *VaultService) Delete(ctx context.Context, userID, vaultID string, force bool) error {
	owns, err := s.vaultRepo.BelongsToUser(ctx, vaultID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return fmt.Errorf("%w: vault not found", model.ErrNotFound)
	}
	return s.vaultRepo.Delete(ctx, vaultID)
}
