package service

import (
	"context"
	"errors"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service/dto"
	"key-haven-back/pkg/helper"
	"time"
)

var (
	ErrCredentialNotFound    = errors.New("credential not found")
	ErrEncryptionFailed      = errors.New("failed to encrypt password")
	ErrDecryptionFailed      = errors.New("failed to decrypt password")
	ErrInvalidVaultSpecified = errors.New("invalid vault specified")
)

// CredentialService defines the interface for credential-related operations
type CredentialService interface {
	CreateCredential(ctx context.Context, request *dto.CreateCredentialRequest) (*dto.CredentialListItem, error)
	GetCredentialByID(ctx context.Context, id string, masterPassword string) (*dto.CredentialDetailResponse, error)
	GetAllCredentialsByVaultID(ctx context.Context, vaultID string) ([]*dto.CredentialListItem, error)
	UpdateCredential(ctx context.Context, request *dto.UpdateCredentialRequest) (*dto.CredentialListItem, error)
	DeleteCredential(ctx context.Context, id string) error
}

type credentialService struct {
	credentialRepo repository.CredentialRepository
	vaultService   VaultService
	vaultRepo      repository.VaultRepository
}

func NewCredentialService(credentialRepo repository.CredentialRepository, vaultService VaultService, vaultRepo repository.VaultRepository) CredentialService {
	return &credentialService{
		credentialRepo: credentialRepo,
		vaultService:   vaultService,
		vaultRepo:      vaultRepo,
	}
}

func (s *credentialService) CreateCredential(ctx context.Context, request *dto.CreateCredentialRequest) (*dto.CredentialListItem, error) {
	var vaultID string
	var err error

	if request.VaultID != "" {
		vault, err := s.vaultRepo.FindByID(ctx, request.VaultID)
		if err != nil {
			if errors.Is(err, repository.ErrVaultNotFound) {
				return nil, ErrInvalidVaultSpecified
			}

			return nil, err
		}

		if vault.UserID != request.UserID {
			return nil, ErrInvalidVaultSpecified
		}

		vaultID = vault.ID
	} else if request.VaultName != "" {
		// Try to find a vault with this name for the user
		vault, err := s.vaultRepo.FindByName(ctx, request.UserID, request.VaultName)
		if err != nil {
			if errors.Is(err, repository.ErrVaultNotFound) {
				// Create a new vault with this name
				newVault, err := s.vaultService.CreateVault(ctx, &dto.CreateVaultRequest{
					UserID:      request.UserID,
					Name:        request.VaultName,
					Description: "Created automatically for new credential",
				})
				if err != nil {
					return nil, err
				}
				vaultID = newVault.ID
			} else {
				return nil, err
			}
		} else {
			vaultID = vault.ID
		}
	} else {
		// Use or create a default vault
		vaultResponse, err := s.vaultService.EnsureDefaultVault(ctx, request.UserID)
		if err != nil {
			return nil, err
		}
		vaultID = vaultResponse.ID
	}

	// Encrypt the password using the master password
	encryptedPassword, err := helper.EncryptPassword(request.Password, request.MasterPassword)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Create the credential
	credential := dto.NewCredential(request, encryptedPassword, vaultID)
	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, err
	}

	return dto.CredentialToListItem(credential), nil
}

// GetCredentialByID retrieves a credential by its ID and decrypts the password
func (s *credentialService) GetCredentialByID(ctx context.Context, id string, masterPassword string) (*dto.CredentialDetailResponse, error) {
	// Find the credential
	credential, err := s.credentialRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	// Get vault info
	vault, err := s.vaultRepo.FindByID(ctx, credential.VaultID)
	if err != nil {
		return nil, err
	}

	// Decrypt the password
	decryptedPassword, err := helper.DecryptPassword(credential.EncryptPassword, masterPassword)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Build the response with vault name and decrypted password
	return &dto.CredentialDetailResponse{
		ID:          credential.ID,
		Name:        credential.Name,
		Description: credential.Description,
		Username:    credential.Username,
		Password:    decryptedPassword,
		URL:         credential.URL,
		VaultID:     credential.VaultID,
		VaultName:   vault.Name,
		CreatedAt:   credential.CreatedAt,
		UpdatedAt:   credential.UpdatedAt,
	}, nil
}

func (s *credentialService) GetAllCredentialsByVaultID(ctx context.Context, vaultID string) ([]*dto.CredentialListItem, error) {
	// Check if the vault exists
	vault, err := s.vaultRepo.FindByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, repository.ErrVaultNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	// Get all credentials in the vault
	credentials, err := s.credentialRepo.FindAllByVaultID(ctx, vaultID)
	if err != nil {
		return nil, err
	}

	// Convert to list items and add vault name
	items := dto.CredentialsToListItems(credentials)
	for _, item := range items {
		item.VaultName = vault.Name
	}

	return items, nil
}

func (s *credentialService) UpdateCredential(ctx context.Context, request *dto.UpdateCredentialRequest) (*dto.CredentialListItem, error) {
	// Check if credential exists
	credential, err := s.credentialRepo.FindByID(ctx, request.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	// Update basic fields
	credential.Name = request.Name
	credential.Description = request.Description
	credential.Username = request.Username
	credential.URL = request.URL
	credential.UpdatedAt = time.Now()

	// Update password if provided
	if request.Password != "" {
		encryptedPassword, err := helper.EncryptPassword(request.Password, request.MasterPassword)
		if err != nil {
			return nil, ErrEncryptionFailed
		}
		credential.EncryptPassword = encryptedPassword
	}

	// Save the updates
	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, err
	}

	return dto.CredentialToListItem(credential), nil
}

// DeleteCredential deletes a credential
func (s *credentialService) DeleteCredential(ctx context.Context, id string) error {
	// Check if credential exists
	_, err := s.credentialRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return ErrCredentialNotFound
		}
		return err
	}

	// Delete the credential
	return s.credentialRepo.Delete(ctx, id)
}
