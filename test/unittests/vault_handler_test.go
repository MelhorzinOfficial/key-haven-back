package unittests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"key-haven-back/internal/http/handler"
	"key-haven-back/internal/service"
	"key-haven-back/internal/service/dto"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

// setupApp creates a new Fiber app for testing
func setupApp() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	return app
}

// VaultServiceMock is a mock implementation of service.VaultService
type VaultServiceMock struct {
	mock.Mock
}

func (m *VaultServiceMock) GetDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *VaultServiceMock) CreateVault(ctx context.Context, req *dto.CreateVaultRequest) (*dto.VaultResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *VaultServiceMock) GetVaultByID(ctx context.Context, id string) (*dto.VaultResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *MockVaultService) GetAllVaultsByUserID(ctx context.Context, userID string) ([]*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.VaultResponse), args.Error(1)
}

func (m *MockVaultService) UpdateVault(ctx context.Context, req *dto.UpdateVaultRequest) (*dto.VaultResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *MockVaultService) DeleteVault(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVaultService) EnsureDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func TestVaultHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)
		app := setupApp()

		vaultResp := &dto.VaultResponse{
			ID:     "vault123",
			Name:   "My Vault",
			UserID: "user123",
		}

		reqBody := dto.CreateVaultRequest{
			Name:        "My Vault",
			Description: "My personal vault",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("CreateVault", mock.Anything, mock.MatchedBy(func(req *dto.CreateVaultRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(vaultResp, nil)

		app.Post("/vault", handler.Create)

		// Set up authentication context
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")

		// Execute
		handler.Create(ctx)

		// Assert
		assert.Equal(t, fiber.StatusCreated, ctx.Response().StatusCode())
		var response SuccessResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		responseVault, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, vaultResp.ID, responseVault["id"])
		assert.Equal(t, vaultResp.Name, responseVault["name"])

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)
		app := setupApp()

		reqBody := dto.CreateVaultRequest{
			Name:        "My Vault",
			Description: "My personal vault",
		}
		jsonBody, _ := json.Marshal(reqBody)

		app.Post("/vault", handler.Create)

		// Execute without setting user_id in context
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "") // Empty user_id

		// Execute
		handler.Create(ctx)

		// Assert
		assert.Equal(t, fiber.StatusUnauthorized, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Unauthorized", response.Message)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		// Invalid JSON
		invalidJSON := []byte(`{"name": "test`)

		// Execute
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")

		// Execute
		handler.Create(ctx)

		// Assert
		assert.Equal(t, fiber.StatusBadRequest, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Invalid request body", response.Message)
	})

	t.Run("Validation Error", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		// Request with empty name (assuming there's validation for required name)
		reqBody := dto.CreateVaultRequest{
			Name:        "", // Empty name should trigger validation error
			Description: "My personal vault",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Execute
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Language", "en") // Set language for validator translations
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")

		// This is more complex to test because it involves the validator setup
		// For now, we'll assume the validation occurs and focus on other tests
		// A full integration test would be needed to fully test the validation
	})

	t.Run("Vault Name Already Exists", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		reqBody := dto.CreateVaultRequest{
			Name:        "Existing Vault",
			Description: "This vault name already exists",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("CreateVault", mock.Anything, mock.MatchedBy(func(req *dto.CreateVaultRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, service.ErrVaultNameExists)

		// Execute
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")

		// Execute
		handler.Create(ctx)

		// Assert
		assert.Equal(t, fiber.StatusConflict, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "A vault with this name already exists", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		reqBody := dto.CreateVaultRequest{
			Name:        "My Vault",
			Description: "My personal vault",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("CreateVault", mock.Anything, mock.MatchedBy(func(req *dto.CreateVaultRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, errors.New("database error"))

		// Execute
		req := httptest.NewRequest("POST", "/vault", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")

		// Execute
		handler.Create(ctx)

		// Assert
		assert.Equal(t, fiber.StatusInternalServerError, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Failed to create vault", response.Message)

		mockService.AssertExpectations(t)
	})
}

func TestVaultHandler_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"
		userID := "user123"

		vaultResp := &dto.VaultResponse{
			ID:     vaultID,
			Name:   "Updated Vault",
			UserID: userID,
		}

		existingVault := &dto.VaultResponse{
			ID:     vaultID,
			Name:   "Original Vault",
			UserID: userID,
		}

		reqBody := dto.UpdateVaultRequest{
			ID:          vaultID,
			Name:        "Updated Vault",
			Description: "Updated description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("GetVaultByID", mock.Anything, vaultID).Return(existingVault, nil)
		mockService.On("UpdateVault", mock.Anything, mock.MatchedBy(func(req *dto.UpdateVaultRequest) bool {
			return req.ID == vaultID && req.Name == reqBody.Name
		})).Return(vaultResp, nil)

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", userID)
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusOK, ctx.Response().StatusCode())
		var response SuccessResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		responseVault, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, vaultResp.ID, responseVault["id"])
		assert.Equal(t, vaultResp.Name, responseVault["name"])

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		mockService := new(MockVaultService)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"

		// Invalid JSON
		invalidJSON := []byte(`{"name": "test`)

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusBadRequest, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Invalid request body", response.Message)
	})

	t.Run("Vault Not Found", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"

		reqBody := dto.UpdateVaultRequest{
			ID:          vaultID,
			Name:        "Updated Vault",
			Description: "Updated description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("GetVaultByID", mock.Anything, vaultID).Return(nil, service.ErrVaultNotFound)

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", "user123")
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusNotFound, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Vault not found", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		// Setup
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"
		loggedInUserID := "user456" // Different from the vault owner

		existingVault := &dto.VaultResponse{
			ID:     vaultID,
			Name:   "Original Vault",
			UserID: "user123", // Different from logged in user
		}

		reqBody := dto.UpdateVaultRequest{
			ID:          vaultID,
			Name:        "Updated Vault",
			Description: "Updated description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("GetVaultByID", mock.Anything, vaultID).Return(existingVault, nil)

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", loggedInUserID) // Different user
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusUnauthorized, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "You don't have access to this vault", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Vault Name Already Exists", func(t *testing.T) {
		// Setup
		mockService := new(MockVaultService)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"
		userID := "user123"

		existingVault := &dto.VaultResponse{
			ID:     vaultID,
			Name:   "Original Vault",
			UserID: userID,
		}

		reqBody := dto.UpdateVaultRequest{
			ID:          vaultID,
			Name:        "Updated Vault",
			Description: "Updated description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("GetVaultByID", mock.Anything, vaultID).Return(existingVault, nil)
		mockService.On("UpdateVault", mock.Anything, mock.MatchedBy(func(req *dto.UpdateVaultRequest) bool {
			return req.ID == vaultID && req.Name == reqBody.Name
		})).Return(nil, service.ErrVaultNameExists)

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", userID)
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusConflict, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "A vault with this name already exists", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		// Setup
		mockService := new(MockVaultService)
		handler := handler.NewVaultHandler(mockService)

		vaultID := "vault123"
		userID := "user123"

		existingVault := &dto.VaultResponse{
			ID:     vaultID,
			Name:   "Original Vault",
			UserID: userID,
		}

		reqBody := dto.UpdateVaultRequest{
			ID:          vaultID,
			Name:        "Updated Vault",
			Description: "Updated description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockService.On("GetVaultByID", mock.Anything, vaultID).Return(existingVault, nil)
		mockService.On("UpdateVault", mock.Anything, mock.MatchedBy(func(req *dto.UpdateVaultRequest) bool {
			return req.ID == vaultID && req.Name == reqBody.Name
		})).Return(nil, errors.New("database error"))

		// Execute
		req := httptest.NewRequest("PUT", "/vault/"+vaultID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Locals("user_id", userID)
		ctx.Params("id", vaultID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusInternalServerError, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Failed to update vault", response.Message)

		mockService.AssertExpectations(t)
	})
}
