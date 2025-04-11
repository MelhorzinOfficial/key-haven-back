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
  var tokenJwt string
  var tokenSource string
  var originalLength int

  tokenJwt = c.Cookies("token")
  if tokenJwt != "" {
    tokenSource = "cookie"
    originalLength = len(tokenJwt)
    
    // Check for cookie size limitations
    if originalLength >= 4000 {
      log.Printf("WARNING: Cookie token is very large (%d bytes) and might be truncated by the browser", originalLength)
    }

    // Cookie values are URL-encoded, so decode
    decodedToken, err := url.QueryUnescape(tokenJwt)
    if err == nil && decodedToken != tokenJwt {
      log.Printf("Cookie token was URL-encoded, decoded successfully (before: %d, after: %d bytes)",
        len(tokenJwt), len(decodedToken))
      tokenJwt = decodedToken
    }

    // Also check for any plus signs that might have been converted to spaces
    if strings.Contains(tokenJwt, " ") {
      log.Printf("Token contains spaces, attempting to replace with plus signs")
      tokenJwt = strings.ReplaceAll(tokenJwt, " ", "+")
    }
  }

  // If not in cookies, check Authorization header
  if tokenJwt == "" {
    authHeader := c.Get("Authorization")
    if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
      tokenJwt = strings.TrimSpace(authHeader[7:])
      tokenSource = "authorization header"
      originalLength = len(tokenJwt)
      log.Printf("Found token in authorization header (length: %d)", originalLength)
    }
  }

  if tokenJwt == "" {
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
      "message": "Unauthorized: No authentication token provided",
    })
  }

  // Clean the token - remove any unwanted characters
  tokenJwt = strings.TrimSpace(tokenJwt)

  // Log token format
  if len(tokenJwt) >= 15 {
    log.Printf("Token prefix: %s...", tokenJwt[:15])
  }

  // Determine if it's a proper PASETO token (should start with v4.local.)
  if !strings.HasPrefix(tokenJwt, "v4.local.") {
    log.Printf("Token does not have the expected PASETO prefix 'v4.local.'")
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
      "message": "Invalid token format: not a valid PASETO token",
    })
  }

  // Check for common encoding issues
  parts := strings.Split(tokenJwt, ".")
  if len(parts) != 3 {
    log.Printf("Invalid token format: expected 3 parts separated by dots, got %d", len(parts))
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
      "message": "Invalid token format: wrong number of segments",
    })
  }

  // Verify the base64 segment can be decoded
  payload := parts[2]
  if _, err := base64.RawURLEncoding.DecodeString(payload); err != nil {
    log.Printf("Base64 decoding failed: %v", err)

    // Try to find where the Base64 error occurs
    chunk := 100
    for i := 0; i < len(payload); i += chunk {
      end := i + chunk
      if end > len(payload) {
        end = len(payload)
      }
      segment := payload[i:end]
      _, err := base64.RawURLEncoding.DecodeString(segment)
      if err != nil {
        log.Printf("Base64 error in segment starting at position %d: %v", i, err)
        break
      }
    }

    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
      "message": "Invalid token encoding",
      "error":   "The token contains invalid base64 data",
      "detail":  err.Error(),
    })
  }

  // Validate token
  claims, err := secret.ValidateToken(tokenJwt)
  if err != nil {
    // Log the error to help with debugging
    log.Printf("Token validation error (from %s): %v", tokenSource, err)

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
