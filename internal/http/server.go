package http

import (
	"context"
	"fmt"
	"key-haven-back/config"
	handler "key-haven-back/internal/http/handler"
	"key-haven-back/internal/http/router"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/fx"

	docsPkg "key-haven-back/pkg/docs"
	errorHandler "key-haven-back/pkg/error"
	"key-haven-back/pkg/validator"
)

func StartServer(lc fx.Lifecycle, app *fiber.App, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := app.Listen(fmt.Sprintf(":%s", cfg.Port)); err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})
}

func NewServer(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	credentialHandler *handler.CredentialHandler,
	vaultHandler *handler.VaultHandler,
	registerRoutes router.RegisterRoutesFunc,
	registerSwagger router.RegisterSwaggerRoutesFunc,
	registerDocs docsPkg.RegisterDocsRouterFunc,
) *fiber.App {

	app := fiber.New(
		fiber.Config{
			ErrorHandler:    errorHandler.GlobalErrorHandler,
			StructValidator: validator.NewStructValidator(),
		},
	)

	// Configure middlewares
	app.Use(cors.New())
	app.Use(requestid.New())
	app.Use(recover.New())

	// Health check route
	app.Get("/", func(c fiber.Ctx) error {
		var data = fiber.Map{"status": "ok", "message": "Welcome to Key Haven API"}
		return c.Status(fiber.StatusOK).JSON(data)
	})

	// Register API routes
	registerRoutes(app, authHandler, credentialHandler, vaultHandler)
	registerSwagger(app)
	registerDocs(app)

	return app
}
