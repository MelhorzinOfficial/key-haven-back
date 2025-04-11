package unittests

import (
	"context"
	"encoding/json"
	"io"
	"key-haven-back/internal/http/handler"
	"key-haven-back/internal/service"
	"key-haven-back/internal/service/dto"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock VaultService that matches the correct interface signature
type VaultServiceMock struct {
	mock.Mock
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

func (m *VaultServiceMock) GetAllVaultsByUserID(ctx context.Context, userID string) ([]*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.VaultResponse), args.Error(1)
}

func (m *VaultServiceMock) UpdateVault(ctx context.Context, req *dto.UpdateVaultRequest) (*dto.VaultResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *VaultServiceMock) DeleteVault(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *VaultServiceMock) GetDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *VaultServiceMock) EnsureDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

// ErrorResponse is already defined in another file in the package
// SuccessResponse is already defined in another file in the package

// Simple test for vault validation
func TestVaultCreateValidation(t *testing.T) {
	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		// Register the route
		app.Post("/vault", func(c fiber.Ctx) error {
			// Set the user_id in context
			c.Locals("user_id", "user123")
			return handler.Create(c)
		})

		// Create a test request with invalid JSON
		req := httptest.NewRequest("POST", "/vault", strings.NewReader(`{"name": "Test Vault"`)) // Invalid JSON
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Check status code
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Read response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// Parse error response
		var errorResp ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", errorResp.Message)
	})

	t.Run("Vault Name Already Exists", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		// Register the route with middleware to set user_id
		app.Post("/vault", func(c fiber.Ctx) error {
			c.Locals("user_id", "user123")
			return handler.Create(c)
		})

		// Create valid request body
		reqBody := dto.CreateVaultRequest{
			Name:        "Existing Vault",
			Description: "This vault name already exists",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock the service to return ErrVaultNameExists
		mockService.On("CreateVault", mock.Anything, mock.MatchedBy(func(req *dto.CreateVaultRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, service.ErrVaultNameExists)

		// Create test request
		req := httptest.NewRequest("POST", "/vault", strings.NewReader(string(jsonBody)))
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Check status code
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		// Read response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// Parse error response
		var errorResp ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "A vault with this name already exists", errorResp.Message)

		// Verify mock expectations
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockService := new(VaultServiceMock)
		handler := handler.NewVaultHandler(mockService)

		// Register the route, but don't set user_id in context
		app.Post("/vault", handler.Create)

		// Create valid request body
		reqBody := dto.CreateVaultRequest{
			Name:        "Test Vault",
			Description: "Test Description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Create test request
		req := httptest.NewRequest("POST", "/vault", strings.NewReader(string(jsonBody)))
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Check status code
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		// Read response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// Parse error response
		var errorResp ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized", errorResp.Message)
	})
}
