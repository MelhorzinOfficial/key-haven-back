package middleware

import (
	"encoding/base64"
	"errors"
	"key-haven-back/pkg/secret"
	"log"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func IsAuthenticatedHandler(c fiber.Ctx) error {
	tokenJwt, err := checkToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: No authentication token provided",
			"error":   err.Error(),
		})
	}

	// Validate token
	claims, err := secret.ValidateToken(tokenJwt)
	if err != nil {
		if errors.Is(err, secret.ErrTokenExpired) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token expired",
			})
		} else if errors.Is(err, secret.ErrEmptyToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Empty token provided",
			})
		} else if errors.Is(err, secret.ErrInvalidToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authentication failed",
			"error":   err.Error(),
		})
	}

	// Store claims in context for use in route handlers
	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)

	return c.Next()
}

func checkToken(c fiber.Ctx) (string, error) {
	tokenJwt, err := checkCookieToken(c.Cookies("token"))
	if err == nil && tokenJwt != "" {
		return tokenJwt, nil
	}

	tokenJwt, err = checkAuthToken(c.Get("Authorization"))
	if err == nil && tokenJwt != "" {
		return tokenJwt, nil
	}

	return "", errors.New("no authentication token provided")
}

func checkCookieToken(tokenJwt string) (string, error) {
	if tokenJwt != "" {
		originalLength := len(tokenJwt)

		if originalLength >= 4000 {
			log.Printf("WARNING: Cookie token is very large (%d bytes) and might be truncated by the browser", originalLength)
		}

		decodedToken, err := url.QueryUnescape(tokenJwt)
		if err == nil && decodedToken != tokenJwt {
			log.Printf("Cookie token was URL-encoded, decoded successfully (before: %d, after: %d bytes)",
				len(tokenJwt), len(decodedToken))
			tokenJwt = decodedToken
		}

		if strings.Contains(tokenJwt, " ") {
			log.Printf("Token contains spaces, attempting to replace with plus signs")
			tokenJwt = strings.ReplaceAll(tokenJwt, " ", "+")
		}

		if validateTokenFormat(tokenJwt) {
			return tokenJwt, nil
		}
	}

	return "", errors.New("invalid token format in cookie")
}

func checkAuthToken(authHeader string) (string, error) {
	tokenJwt := strings.TrimPrefix(authHeader, "Bearer ")
	if !validateTokenFormat(tokenJwt) {
		return "", errors.New("invalid token format in authorization header")
	}
	return tokenJwt, nil
}

func validateTokenFormat(token string) bool {
	token = strings.TrimSpace(token)

	if !strings.HasPrefix(token, "v4.local.") {
		return false
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	payload := parts[2]
	_, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		log.Printf("Base64 decoding failed: %v", err)
		return false
	}

	return true
}
