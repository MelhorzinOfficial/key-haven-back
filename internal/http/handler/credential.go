package handler

import (
	"errors"
	"key-haven-back/internal/http/response"
	"key-haven-back/internal/service"
	"key-haven-back/internal/service/dto"

	"github.com/gofiber/fiber/v3"
)

type CredentialHandler struct {
	credentialService service.CredentialService
	vaultService      service.VaultService
}

func NewCredentialHandler(credentialService service.CredentialService, vaultService service.VaultService) *CredentialHandler {
	return &CredentialHandler{
		credentialService: credentialService,
		vaultService:      vaultService,
	}
}

// Create godoc
// @Summary Create a new credential
// @Description Creates a new credential optionally in a specified vault
// @Tags Credentials
// @Accept json
// @Produce json
// @Param request body dto.CreateCredentialRequest true "Credential creation data"
// @Success 201 {object} dto.CredentialListItem
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /credential [post]
// @Security ApiKeyAuth
func (h *CredentialHandler) Create(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}

	userID := c.Locals("user_id").(string)
	if userID == "" {
		return res.Message(fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.CreateCredentialRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	req.UserID = userID

	credential, err := h.credentialService.CreateCredential(c.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidVaultSpecified) {
			return res.Message(fiber.StatusBadRequest, "Invalid vault specified")
		}
		if errors.Is(err, service.ErrEncryptionFailed) {
			return res.Message(fiber.StatusBadRequest, "Failed to encrypt credential password")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to create credential")
	}

	return res.Created(SuccessResponse{Data: credential})
}

// GetByID godoc
// @Summary Get credential details
// @Description Retrieves a credential by its ID and decrypts the password
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path string true "Credential ID"
// @Param master_password query string true "Master password for decryption"
// @Success 200 {object} dto.CredentialDetailResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /credential/{id} [get]
// @Security ApiKeyAuth
func (h *CredentialHandler) GetByID(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")
	masterPassword := c.Query("master_password")

	if masterPassword == "" {
		return res.Message(fiber.StatusBadRequest, "Master password is required")
	}

	credential, err := h.credentialService.GetCredentialByID(c.Context(), id, masterPassword)
	if err != nil {
		if errors.Is(err, service.ErrCredentialNotFound) {
			return res.Message(fiber.StatusNotFound, "Credential not found")
		}
		if errors.Is(err, service.ErrDecryptionFailed) {
			return res.Message(fiber.StatusBadRequest, "Failed to decrypt password. Invalid master password?")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve credential")
	}

	return res.Ok(SuccessResponse{Data: credential})
}

// GetAllByVault godoc
// @Summary Get all credentials in a vault
// @Description Retrieves all credentials in a specified vault
// @Tags Credentials
// @Accept json
// @Produce json
// @Param vault_id path string true "Vault ID"
// @Success 200 {array} dto.CredentialListItem
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Vault not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /credential/vault/{vault_id} [get]
// @Security ApiKeyAuth
func (h *CredentialHandler) GetAllByVault(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	vaultID := c.Params("vault_id")

	credentials, err := h.credentialService.GetAllCredentialsByVaultID(c.Context(), vaultID)
	if err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve credentials")
	}

	return res.Ok(SuccessResponse{Data: credentials})
}

// Update godoc
// @Summary Update a credential
// @Description Updates a credential's details and optionally its password
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path string true "Credential ID"
// @Param request body dto.UpdateCredentialRequest true "Credential update data"
// @Success 200 {object} dto.CredentialListItem
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /credential/{id} [put]
// @Security ApiKeyAuth
func (h *CredentialHandler) Update(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")

	var req dto.UpdateCredentialRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	req.ID = id

	credential, err := h.credentialService.UpdateCredential(c.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCredentialNotFound) {
			return res.Message(fiber.StatusNotFound, "Credential not found")
		}
		if errors.Is(err, service.ErrEncryptionFailed) {
			return res.Message(fiber.StatusBadRequest, "Failed to encrypt credential password")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to update credential")
	}

	return res.Ok(SuccessResponse{Data: credential})
}

// Delete godoc
// @Summary Delete a credential
// @Description Deletes a credential by its ID
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path string true "Credential ID"
// @Success 200 {object} SuccessResponse "Success message"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /credential/{id} [delete]
// @Security ApiKeyAuth
func (h *CredentialHandler) Delete(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")

	if err := h.credentialService.DeleteCredential(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrCredentialNotFound) {
			return res.Message(fiber.StatusNotFound, "Credential not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to delete credential")
	}

	return res.Ok(SuccessResponse{Data: "Credential deleted successfully"})
}

func (h *CredentialHandler) GetDefaultVault(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}

	userID := c.Locals("user_id").(string)
	if userID == "" {
		return res.Message(fiber.StatusUnauthorized, "Unauthorized")
	}

	vault, err := h.vaultService.EnsureDefaultVault(c.Context(), userID)
	if err != nil {
		return res.Message(fiber.StatusInternalServerError, "Failed to get default vault")
	}

	return res.Ok(SuccessResponse{Data: vault})
}
