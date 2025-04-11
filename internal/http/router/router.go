package router

import (
	handler "key-haven-back/internal/http/handler"

	"github.com/gofiber/fiber/v3"
)

type RegisterRoutesFunc func(app *fiber.App, authHandler *handler.AuthHandler, credentialHandler *handler.CredentialHandler, vaultHandler *handler.VaultHandler)
type RegisterSwaggerRoutesFunc func(app *fiber.App)

func RegisterRoutesFuncProvider() RegisterRoutesFunc {
	return RegisterRoutes
}

func RegisterSwaggerRoutesFuncProvider() RegisterSwaggerRoutesFunc {
	return RegisterSwaggerRoutes
}
