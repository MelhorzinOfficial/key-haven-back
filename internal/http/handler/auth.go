package handler

import (
	"errors"
	"key-haven-back/internal/http/response"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service"
	"key-haven-back/internal/service/dto"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func handleError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{Message: message})
}

func setAuthCookie(c fiber.Ctx, token string, duration time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(duration),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User registration data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 409 {object} ErrorResponse "Email already in use"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var res = response.HTTPResponse{Ctx: c}

	var req dto.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	user, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyUsed) {
			return res.Message(fiber.StatusConflict, "Email already in use")
		}
		return res.Message(fiber.StatusInternalServerError, "Failed to process registration")
	}

	return res.Created(SuccessResponse{Data: user})
}

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns a token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User login data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid email or password"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return handleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	r, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return handleError(c, fiber.StatusUnauthorized, "Invalid email or password")
		}
		return handleError(c, fiber.StatusInternalServerError, "Failed to process login")
	}

	setAuthCookie(c, r.Token, 24*time.Hour)
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{Data: r})
}

// Logout godoc
// @Summary User logout
// @Description Logs out the user by clearing the authentication cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	setAuthCookie(c, "", -time.Hour) // Expire the cookie
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Data: "Logged out successfully",
	})
}
