package dto

import (
	"key-haven-back/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

type CreateVaultRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateVaultRequest struct {
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type VaultResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewVault creates a new Vault object from a CreateVaultRequest
func NewVault(req *CreateVaultRequest) *user.Vault {
	now := time.Now()
	return &user.Vault{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// VaultToResponse converts a Vault domain object to a VaultResponse DTO
func VaultToResponse(vault *user.Vault) *VaultResponse {
	return &VaultResponse{
		ID:          vault.ID,
		UserID:      vault.UserID,
		Name:        vault.Name,
		Description: vault.Description,
		CreatedAt:   vault.CreatedAt,
		UpdatedAt:   vault.UpdatedAt,
	}
}

// VaultsToResponses converts a slice of Vault domain objects to VaultResponse DTOs
func VaultsToResponses(vaults []*user.Vault) []*VaultResponse {
	responses := make([]*VaultResponse, len(vaults))
	for i, vault := range vaults {
		responses[i] = VaultToResponse(vault)
	}
	return responses
}
