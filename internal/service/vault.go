package service

import (
	"context"
	"errors"
	"key-haven-back/internal/domain/user"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service/dto"
	"log"
	"time"
)

var (
	ErrVaultNotFound   = errors.New("vault not found")
	ErrVaultNameExists = errors.New("vault name already exists for this user")
)

type VaultService interface {
	CreateVault(ctx context.Context, request *dto.CreateVaultRequest) (*dto.VaultResponse, error)
	GetVaultByID(ctx context.Context, id string) (*dto.VaultResponse, error)
	GetDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error)
	GetAllVaultsByUserID(ctx context.Context, userID string) ([]*dto.VaultResponse, error)
	UpdateVault(ctx context.Context, request *dto.UpdateVaultRequest) (*dto.VaultResponse, error)
	DeleteVault(ctx context.Context, id string) error
	EnsureDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error)
}

type vaultService struct {
	vaultRepo repository.VaultRepository
}

func NewVaultService(vaultRepo repository.VaultRepository) VaultService {
	return &vaultService{
		vaultRepo: vaultRepo,
	}
}

func (s *vaultService) CreateVault(ctx context.Context, request *dto.CreateVaultRequest) (*dto.VaultResponse, error) {
	vault := dto.NewVault(request)

	if err := s.vaultRepo.Create(ctx, vault); err != nil {
		return nil, err
	}

	return dto.VaultToResponse(vault), nil
}

func (s *vaultService) GetVaultByID(ctx context.Context, id string) (*dto.VaultResponse, error) {
	vault, err := s.vaultRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrVaultNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	return dto.VaultToResponse(vault), nil
}

func (s *vaultService) GetDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	vault, err := s.vaultRepo.FindDefaultByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrVaultNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	return dto.VaultToResponse(vault), nil
}

func (s *vaultService) GetAllVaultsByUserID(ctx context.Context, userID string) ([]*dto.VaultResponse, error) {
	vaults, err := s.vaultRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return dto.VaultsToResponses(vaults), nil
}

func (s *vaultService) UpdateVault(ctx context.Context, request *dto.UpdateVaultRequest) (*dto.VaultResponse, error) {
	vault, err := s.vaultRepo.FindByID(ctx, request.ID)
	if err != nil {
		if errors.Is(err, repository.ErrVaultNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	vault.Name = request.Name
	vault.Description = request.Description
	vault.UpdatedAt = time.Now()

	if err := s.vaultRepo.Update(ctx, vault); err != nil {
		return nil, err
	}

	return dto.VaultToResponse(vault), nil
}

func (s *vaultService) DeleteVault(ctx context.Context, id string) error {
	_, err := s.vaultRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrVaultNotFound) {
			return ErrVaultNotFound
		}
		return err
	}

	// TODO: In a real app, we'd need to handle the credentials in this vault
	// Either move them to another vault or delete them

	return s.vaultRepo.Delete(ctx, id)
}

func (s *vaultService) EnsureDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	vault, err := s.vaultRepo.FindDefaultByUserID(ctx, userID)
	if err == nil && vault != nil {

		return dto.VaultToResponse(vault), nil
	}

	if err != nil && !errors.Is(err, repository.ErrVaultNotFound) {
		return nil, err
	}

	now := time.Now()
	defaultVault := &user.Vault{
		UserID:      userID,
		Name:        "Default",
		Description: "Default vault for storing credentials",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.vaultRepo.Create(ctx, defaultVault); err != nil {
		log.Printf("Failed to create default vault for user %s: %v", userID, err)
		return nil, err
	}

	return dto.VaultToResponse(defaultVault), nil
}
