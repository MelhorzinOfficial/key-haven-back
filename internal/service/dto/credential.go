package dto

import (
	"key-haven-back/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

type CreateCredentialRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	VaultID        string `json:"vault_id"`
	VaultName      string `json:"vault_name"`
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description"`
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required,min=12"`
	URL            string `json:"url"`
	MasterPassword string `json:"master_password" validate:"required,min=12"`
}

type UpdateCredentialRequest struct {
	ID             string `json:"id" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description"`
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password"`
	URL            string `json:"url"`
	MasterPassword string `json:"master_password" validate:"required,min=12"`
}

type CredentialListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Username    string    `json:"username"`
	URL         string    `json:"url"`
	VaultID     string    `json:"vault_id"`
	VaultName   string    `json:"vault_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CredentialDetailResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	URL         string    `json:"url"`
	VaultID     string    `json:"vault_id"`
	VaultName   string    `json:"vault_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewCredential(req *CreateCredentialRequest, encryptedPassword string, vaultID string) *user.Credential {
	now := time.Now()
	return &user.Credential{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		Username:        req.Username,
		EncryptPassword: encryptedPassword,
		URL:             req.URL,
		VaultID:         vaultID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// CredentialToListItem converts a Credential domain object to a CredentialListItem DTO (without password)
func CredentialToListItem(credential *user.Credential) *CredentialListItem {
	return &CredentialListItem{
		ID:          credential.ID,
		Name:        credential.Name,
		Description: credential.Description,
		Username:    credential.Username,
		URL:         credential.URL,
		VaultID:     credential.VaultID,
		CreatedAt:   credential.CreatedAt,
		UpdatedAt:   credential.UpdatedAt,
	}
}

// CredentialsToListItems converts multiple Credential domain objects to CredentialListItem DTOs
func CredentialsToListItems(credentials []*user.Credential) []*CredentialListItem {
	items := make([]*CredentialListItem, len(credentials))
	for i, cred := range credentials {
		items[i] = CredentialToListItem(cred)
	}
	return items
}
