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

// Mock CredentialService
type CredentialServiceMock struct {
	mock.Mock
}

func (m *CredentialServiceMock) CreateCredential(ctx context.Context, req *dto.CreateCredentialRequest) (*dto.CredentialListItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialListItem), args.Error(1)
}

func (m *CredentialServiceMock) GetCredentialByID(ctx context.Context, id string, masterPassword string) (*dto.CredentialDetailResponse, error) {
	args := m.Called(ctx, id, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialDetailResponse), args.Error(1)
}

func (m *CredentialServiceMock) GetAllCredentialsByVaultID(ctx context.Context, vaultID string) ([]*dto.CredentialListItem, error) {
	args := m.Called(ctx, vaultID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.CredentialListItem), args.Error(1)
}

func (m *CredentialServiceMock) UpdateCredential(ctx context.Context, req *dto.UpdateCredentialRequest) (*dto.CredentialListItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialListItem), args.Error(1)
}

func (m *CredentialServiceMock) DeleteCredential(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Simple test for credential validation
func TestCredentialCreateValidation(t *testing.T) {
	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock) // Reusing the mock from vault_validation_test.go
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route
		app.Post("/credential", func(c fiber.Ctx) error {
			// Set the user_id in context
			c.Locals("user_id", "user123")
			return handler.Create(c)
		})

		// Create a test request with invalid JSON
		req := httptest.NewRequest("POST", "/credential", strings.NewReader(`{"name": "Test Credential"`)) // Invalid JSON
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

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route - but don't set user_id
		app.Post("/credential", handler.Create)

		// Create valid request body
		reqBody := dto.CreateCredentialRequest{
			Name:           "Test Credential",
			Username:       "username",
			Password:       "password123",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Create test request
		req := httptest.NewRequest("POST", "/credential", strings.NewReader(string(jsonBody)))
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

	t.Run("Invalid Vault Specified", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route with middleware to set user_id
		app.Post("/credential", func(c fiber.Ctx) error {
			c.Locals("user_id", "user123")
			return handler.Create(c)
		})

		// Create valid request body
		reqBody := dto.CreateCredentialRequest{
			Name:           "Test Credential",
			Username:       "username",
			Password:       "password123",
			VaultID:        "invalid-vault-id",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock the service to return ErrInvalidVaultSpecified
		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123" && req.VaultID == "invalid-vault-id"
		})).Return(nil, service.ErrInvalidVaultSpecified)

		// Create test request
		req := httptest.NewRequest("POST", "/credential", strings.NewReader(string(jsonBody)))
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
		assert.Equal(t, "Invalid vault specified", errorResp.Message)

		// Verify mock expectations
		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Encryption Failed", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route with middleware to set user_id
		app.Post("/credential", func(c fiber.Ctx) error {
			c.Locals("user_id", "user123")
			return handler.Create(c)
		})

		// Create valid request body
		reqBody := dto.CreateCredentialRequest{
			Name:           "Test Credential",
			Username:       "username",
			Password:       "password123",
			MasterPassword: "invalid-master", // Invalid master password will cause encryption to fail
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock the service to return ErrEncryptionFailed
		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, service.ErrEncryptionFailed)

		// Create test request
		req := httptest.NewRequest("POST", "/credential", strings.NewReader(string(jsonBody)))
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
		assert.Equal(t, "Failed to encrypt credential password", errorResp.Message)

		// Verify mock expectations
		mockCredentialService.AssertExpectations(t)
	})
}

func TestCredentialUpdateValidation(t *testing.T) {
	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route
		app.Put("/credential/:id", handler.Update)

		// Create a test request with invalid JSON
		req := httptest.NewRequest("PUT", "/credential/cred123", strings.NewReader(`{"name": "Test Credential"`)) // Invalid JSON
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

	t.Run("Credential Not Found", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route
		app.Put("/credential/:id", handler.Update)

		credID := "non-existent-cred"

		// Create valid request body
		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock the service to return ErrCredentialNotFound
		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID
		})).Return(nil, service.ErrCredentialNotFound)

		// Create test request
		req := httptest.NewRequest("PUT", "/credential/"+credID, strings.NewReader(string(jsonBody)))
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Check status code
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		// Read response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// Parse error response
		var errorResp ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "Credential not found", errorResp.Message)

		// Verify mock expectations
		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Encryption Failed", func(t *testing.T) {
		// Setup
		app := fiber.New()
		mockCredentialService := new(CredentialServiceMock)
		mockVaultService := new(VaultServiceMock)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Register the route
		app.Put("/credential/:id", handler.Update)

		credID := "cred123"

		// Create valid request body
		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			Password:       "updated-secret123",
			MasterPassword: "invalid-master", // Invalid master password will cause encryption to fail
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock the service to return ErrEncryptionFailed
		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID && req.MasterPassword == "invalid-master"
		})).Return(nil, service.ErrEncryptionFailed)

		// Create test request
		req := httptest.NewRequest("PUT", "/credential/"+credID, strings.NewReader(string(jsonBody)))
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
		assert.Equal(t, "Failed to encrypt credential password", errorResp.Message)

		// Verify mock expectations
		mockCredentialService.AssertExpectations(t)
	})
}
