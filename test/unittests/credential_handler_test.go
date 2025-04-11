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

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Message string `json:"message"`
}

// SuccessResponse represents the structure of success responses
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// MockVaultService is a mock implementation of service.VaultService
type MockVaultService struct {
	mock.Mock
}

func (m *MockVaultService) CreateVault(ctx context.Context, req *dto.CreateVaultRequest) (*dto.VaultResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

func (m *MockVaultService) GetVaultByID(ctx context.Context, id string) (*dto.VaultResponse, error) {
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

func (m *MockVaultService) GetDefaultVault(ctx context.Context, userID string) (*dto.VaultResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultResponse), args.Error(1)
}

// MockCredentialService is a mock implementation of service.CredentialService
type MockCredentialService struct {
	mock.Mock
}

func (m *MockCredentialService) CreateCredential(ctx context.Context, req *dto.CreateCredentialRequest) (*dto.CredentialListItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialListItem), args.Error(1)
}

func (m *MockCredentialService) GetCredentialByID(ctx context.Context, id string, masterPassword string) (*dto.CredentialDetailResponse, error) {
	args := m.Called(ctx, id, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialDetailResponse), args.Error(1)
}

func (m *MockCredentialService) GetAllCredentialsByVaultID(ctx context.Context, vaultID string) ([]*dto.CredentialListItem, error) {
	args := m.Called(ctx, vaultID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.CredentialListItem), args.Error(1)
}

func (m *MockCredentialService) UpdateCredential(ctx context.Context, req *dto.UpdateCredentialRequest) (*dto.CredentialListItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CredentialListItem), args.Error(1)
}

func (m *MockCredentialService) DeleteCredential(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupApp(handler *handler.CredentialHandler) *fiber.App {
	app := fiber.New()
	app.Post("/credential", handler.Create)
	app.Put("/credential/:id", handler.Update)
	return app
}

func TestCredentialHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credentialResp := &dto.CredentialListItem{
			ID:       "cred123",
			Name:     "My Credential",
			Username: "username",
			URL:      "example.com",
			VaultID:  "vault123",
		}

		reqBody := dto.CreateCredentialRequest{
			Name:           "My Credential",
			Username:       "username",
			URL:            "example.com",
			Password:       "secret123",
			VaultID:        "vault123",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(credentialResp, nil)

		// Execute
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(jsonBody))
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

		responseCredential, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, credentialResp.ID, responseCredential["id"])
		assert.Equal(t, credentialResp.Name, responseCredential["name"])

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		reqBody := dto.CreateCredentialRequest{
			Name:           "My Credential",
			Username:       "username",
			URL:            "example.com",
			Password:       "secret123",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Execute without setting user_id in context
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(jsonBody))
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
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		// Invalid JSON
		invalidJSON := []byte(`{"name": "test`)

		// Execute
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(invalidJSON))
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

	t.Run("Invalid Vault Specified", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		reqBody := dto.CreateCredentialRequest{
			Name:           "My Credential",
			Username:       "username",
			URL:            "example.com",
			Password:       "secret123",
			VaultID:        "invalid-vault-id",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123" && req.VaultID == "invalid-vault-id"
		})).Return(nil, service.ErrInvalidVaultSpecified)

		// Execute
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Invalid vault specified", response.Message)

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Encryption Failed", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		reqBody := dto.CreateCredentialRequest{
			Name:           "My Credential",
			Username:       "username",
			URL:            "example.com",
			Password:       "secret123",
			MasterPassword: "invalid-master",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, service.ErrEncryptionFailed)

		// Execute
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Failed to encrypt credential password", response.Message)

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		reqBody := dto.CreateCredentialRequest{
			Name:           "My Credential",
			Username:       "username",
			URL:            "example.com",
			Password:       "secret123",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("CreateCredential", mock.Anything, mock.MatchedBy(func(req *dto.CreateCredentialRequest) bool {
			return req.Name == reqBody.Name && req.UserID == "user123"
		})).Return(nil, errors.New("database error"))

		// Execute
		req := httptest.NewRequest("POST", "/credential", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Failed to create credential", response.Message)

		mockCredentialService.AssertExpectations(t)
	})
}

func TestCredentialHandler_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credID := "cred123"

		credentialResp := &dto.CredentialListItem{
			ID:       credID,
			Name:     "Updated Credential",
			Username: "updated-username",
			URL:      "updated-example.com",
			VaultID:  "vault123",
		}

		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			Password:       "updated-secret123",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID && req.Name == reqBody.Name
		})).Return(credentialResp, nil)

		// Execute
		req := httptest.NewRequest("PUT", "/credential/"+credID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Params("id", credID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusOK, ctx.Response().StatusCode())
		var response SuccessResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		responseCredential, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, credentialResp.ID, responseCredential["id"])
		assert.Equal(t, credentialResp.Name, responseCredential["name"])

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credID := "cred123"

		// Invalid JSON
		invalidJSON := []byte(`{"name": "test`)

		// Execute
		req := httptest.NewRequest("PUT", "/credential/"+credID, bytes.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Params("id", credID)

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

	t.Run("Credential Not Found", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credID := "non-existent-cred"

		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID
		})).Return(nil, service.ErrCredentialNotFound)

		// Execute
		req := httptest.NewRequest("PUT", "/credential/"+credID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Params("id", credID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusNotFound, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Credential not found", response.Message)

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Encryption Failed", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credID := "cred123"

		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			Password:       "updated-secret123",
			MasterPassword: "invalid-master", // This will cause encryption to fail
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID && req.Name == reqBody.Name
		})).Return(nil, service.ErrEncryptionFailed)

		// Execute
		req := httptest.NewRequest("PUT", "/credential/"+credID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Params("id", credID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusBadRequest, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Failed to encrypt credential password", response.Message)

		mockCredentialService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		// Setup
		mockCredentialService := new(MockCredentialService)
		mockVaultService := new(MockVaultService)
		handler := handler.NewCredentialHandler(mockCredentialService, mockVaultService)

		credID := "cred123"

		reqBody := dto.UpdateCredentialRequest{
			ID:             credID,
			Name:           "Updated Credential",
			Username:       "updated-username",
			URL:            "updated-example.com",
			MasterPassword: "master123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockCredentialService.On("UpdateCredential", mock.Anything, mock.MatchedBy(func(req *dto.UpdateCredentialRequest) bool {
			return req.ID == credID && req.Name == reqBody.Name
		})).Return(nil, errors.New("database error"))

		// Execute
		req := httptest.NewRequest("PUT", "/credential/"+credID, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := fiber.NewCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetBody(req.Body)
		ctx.Request().Header.SetContentType("application/json")
		ctx.Params("id", credID)

		// Execute
		handler.Update(ctx)

		// Assert
		assert.Equal(t, fiber.StatusInternalServerError, ctx.Response().StatusCode())
		var response ErrorResponse
		err := json.Unmarshal(ctx.Response().Body(), &response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		assert.Equal(t, "Failed to update credential", response.Message)

		mockCredentialService.AssertExpectations(t)
	})
}
