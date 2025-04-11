package router

import (
  handler "key-haven-back/internal/http/handler"
  "key-haven-back/internal/http/middleware"

  "github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, authHandler *handler.AuthHandler, credentialHandler *handler.CredentialHandler, vaultHandler *handler.VaultHandler) {
  // Auth routes
  auth := app.Group("/auth")
  auth.Post("/register", authHandler.Register)
  auth.Post("/login", authHandler.Login)
  auth.Post("/logout", authHandler.Logout)
  
  // Credential routes - protected by authentication
  credential := app.Group("/credential", middleware.IsAuthenticatedHandler)
  credential.Post("/", credentialHandler.Create)
  credential.Get("/:id", credentialHandler.GetByID)
  credential.Put("/:id", credentialHandler.Update)
  credential.Delete("/:id", credentialHandler.Delete)
  credential.Get("/vault/:vault_id", credentialHandler.GetAllByVault)
  credential.Get("/default-vault", credentialHandler.GetDefaultVault)

  // Vault routes - protected by authentication
  vault := app.Group("/vault", middleware.IsAuthenticatedHandler)
  vault.Post("/", vaultHandler.Create)
  vault.Get("/", vaultHandler.GetAll)
  vault.Get("/default", vaultHandler.GetDefault)
  vault.Get("/:id", vaultHandler.GetByID)
  vault.Put("/:id", vaultHandler.Update)
  vault.Delete("/:id", vaultHandler.Delete)
}
