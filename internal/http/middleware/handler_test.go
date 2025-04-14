package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"key-haven-back/pkg/secret"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsAuthenticatedHandler(t *testing.T) {
	validToken, _ := secret.GenerateToken("123", "test@example.com", time.Hour)

	tests := map[string]struct {
		setup       func(req *http.Request)
		wantStatus  int
		wantMessage string
		wantNext    bool
	}{
		"valid_token_in_header": {
			setup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+validToken)
			},
			wantStatus: fiber.StatusOK,
			wantNext:   true,
		},
		"valid_token_in_cookie": {
			setup: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "token", Value: validToken})
			},
			wantStatus: fiber.StatusOK,
			wantNext:   true,
		},
		"no_token_provided": {
			setup:       func(req *http.Request) {},
			wantStatus:  fiber.StatusUnauthorized,
			wantMessage: "No authentication token provided",
			wantNext:    false,
		},
		"invalid_token": {
			setup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+validToken[1:])
			},
			wantStatus:  fiber.StatusUnauthorized,
			wantMessage: "Invalid token",
			wantNext:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := fiber.New()

			app.Use(IsAuthenticatedHandler)

			app.Get("/", func(c fiber.Ctx) error {
				userID := c.Locals("user_id").(string)
				email := c.Locals("email").(string)
				return c.JSON(fiber.Map{
					"user_id": userID,
					"email":   email,
				})
			})

			req := httptest.NewRequest("GET", "/", nil)
			tt.setup(req)

			resp, err := app.Test(req)
			assert.NoError(t, err, "Error during request test")
			assert.Equal(t, tt.wantStatus, resp.StatusCode, "Status code mismatch")

			if !tt.wantNext {
				respBody := make([]byte, resp.ContentLength)
				_, _ = resp.Body.Read(respBody)
				assert.Contains(t, string(respBody), tt.wantMessage, "Response message mismatch")
			} else {
				respBody := make([]byte, resp.ContentLength)
				_, _ = resp.Body.Read(respBody)
				assert.Contains(t, string(respBody), `"user_id":"123"`, "User ID mismatch")
				assert.Contains(t, string(respBody), `"email":"test@example.com"`, "Email mismatch")
			}
		})
	}
}
