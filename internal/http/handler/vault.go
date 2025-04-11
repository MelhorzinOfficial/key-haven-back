package handler

import (
	"errors"
	"key-haven-back/internal/http/response"
	"key-haven-back/internal/service"
	"key-haven-back/internal/service/dto"

	"github.com/gofiber/fiber/v3"
)

type VaultHandler struct {
	vaultService service.VaultService
}

func NewVaultHandler(vaultService service.VaultService) *VaultHandler {
	return &VaultHandler{
		vaultService: vaultService,
	}
}

// Create godoc
// @Summary Create a new vault
// @Description Creates a new vault for the authenticated user
// @Tags Vaults
// @Accept json
// @Produce json
// @Param request body dto.CreateVaultRequest true "Vault creation data"
// @Success 201 {object} dto.VaultResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 409 {object} ErrorResponse "Vault name already exists"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault [post]
// @Security ApiKeyAuth
func (h *VaultHandler) Create(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}

	// Get the user ID from the authenticated context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return res.Message(fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.CreateVaultRequest
	if err := c.Bind().Body(&req); err != nil {
		return res.Message(fiber.StatusBadRequest, "Invalid request body")
	}

	req.UserID = userID

	vault, err := h.vaultService.CreateVault(c.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrVaultNameExists) {
			return res.Message(fiber.StatusConflict, "A vault with this name already exists")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to create vault")
	}

	return res.Created(SuccessResponse{Data: vault})
}

// GetByID godoc
// @Summary Get vault details
// @Description Retrieves a vault by its ID
// @Tags Vaults
// @Accept json
// @Produce json
// @Param id path string true "Vault ID"
// @Success 200 {object} dto.VaultResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault/{id} [get]
// @Security ApiKeyAuth
func (h *VaultHandler) GetByID(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")

	vault, err := h.vaultService.GetVaultByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve vault")
	}

	userID := c.Locals("user_id").(string)
	if vault.UserID != userID {
		return res.Message(fiber.StatusUnauthorized, "You don't have access to this vault")
	}

	return res.Ok(SuccessResponse{Data: vault})
}

// GetAll godoc
// @Summary Get all vaults
// @Description Retrieves all vaults for the authenticated user
// @Tags Vaults
// @Accept json
// @Produce json
// @Success 200 {array} dto.VaultResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault [get]
// @Security ApiKeyAuth
func (h *VaultHandler) GetAll(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}

	// Get the user ID from the authenticated context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return res.Message(fiber.StatusUnauthorized, "Unauthorized")
	}

	// Get all vaults for the user
	vaults, err := h.vaultService.GetAllVaultsByUserID(c.Context(), userID)
	if err != nil {
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve vaults")
	}

	return res.Ok(SuccessResponse{Data: vaults})
}

// GetDefault godoc
// @Summary Get default vault
// @Description Retrieves or creates the default vault for the authenticated user
// @Tags Vaults
// @Accept json
// @Produce json
// @Success 200 {object} dto.VaultResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault/default [get]
// @Security ApiKeyAuth
func (h *VaultHandler) GetDefault(c fiber.Ctx) error {
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

// Update godoc
// @Summary Update a vault
// @Description Updates a vault's details
// @Tags Vaults
// @Accept json
// @Produce json
// @Param id path string true "Vault ID"
// @Param request body dto.UpdateVaultRequest true "Vault update data"
// @Success 200 {object} dto.VaultResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 409 {object} ErrorResponse "Vault name already exists"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault/{id} [put]
// @Security ApiKeyAuth
func (h *VaultHandler) Update(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")

	var req dto.UpdateVaultRequest
	if err := c.Bind().Body(&req); err != nil {
		return res.Message(fiber.StatusBadRequest, "Invalid request body")
	}

	req.ID = id

	vault, err := h.vaultService.GetVaultByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve vault")
	}

	userID := c.Locals("user_id").(string)
	if vault.UserID != userID {
		return res.Message(fiber.StatusUnauthorized, "You don't have access to this vault")
	}

	updatedVault, err := h.vaultService.UpdateVault(c.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		if errors.Is(err, service.ErrVaultNameExists) {
			return res.Message(fiber.StatusConflict, "A vault with this name already exists")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to update vault")
	}

	return res.Ok(SuccessResponse{Data: updatedVault})
}

// Delete godoc
// @Summary Delete a vault
// @Description Deletes a vault by its ID
// @Tags Vaults
// @Accept json
// @Produce json
// @Param id path string true "Vault ID"
// @Success 200 {object} SuccessResponse "Success message"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /vault/{id} [delete]
// @Security ApiKeyAuth
func (h *VaultHandler) Delete(c fiber.Ctx) error {
	res := response.HTTPResponse{Ctx: c}
	id := c.Params("id")

	vault, err := h.vaultService.GetVaultByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to retrieve vault")
	}

	userID := c.Locals("user_id").(string)
	if vault.UserID != userID {
		return res.Message(fiber.StatusUnauthorized, "You don't have access to this vault")
	}

	// Prevent deletion of the default vault
	if vault.Name == "Default" {
		return res.Message(fiber.StatusBadRequest, "Cannot delete the default vault")
	}

	if err := h.vaultService.DeleteVault(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrVaultNotFound) {
			return res.Message(fiber.StatusNotFound, "Vault not found")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to delete vault")
	}

	return res.Ok(SuccessResponse{Data: "Vault deleted successfully"})
}
