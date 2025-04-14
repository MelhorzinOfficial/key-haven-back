package middleware

import (
	"key-haven-back/pkg/secret"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func IsAuthenticatedHandler(c fiber.Ctx) error {
	var tokenJwt string

	if c.Get("Authorization") != "" {
		tokenJwt = strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	} else if c.Cookies("token") != "" {
		tokenJwt = c.Cookies("token")
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "No authentication token",
		})
	}

	claims, err := secret.ValidateToken(tokenJwt)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
			"error":   err.Error(),
		})
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)

	return c.Next()
}
