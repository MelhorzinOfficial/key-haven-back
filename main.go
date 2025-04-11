package main

import (
	"embed"
	handler "key-haven-back/internal/http/handler"
	"key-haven-back/internal/http/router"
	"log"
	"os"
	"os/exec"

	"key-haven-back/config"
	_ "key-haven-back/docs"
	"key-haven-back/internal/http"
	"key-haven-back/internal/infra/database"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service"
	"key-haven-back/pkg/docs"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

//go:embed docs/swagger.json
//go:embed pkg/docs/*.html
var embeddedFiles embed.FS

func configModule() fx.Option {
	return fx.Provide(config.NewConfig)
}

func databaseModule() fx.Option {
	return fx.Provide(database.NewMongoDBClient)
}

func repositoryModule() fx.Option {
	return fx.Provide(
		func(client database.MongoDBClient) *mongo.Database {
			return client.Database("key-haven")
		},
		repository.NewUserRepository,
		repository.NewVaultRepository,
		repository.NewCredentialRepository,
	)
}

func serviceModule() fx.Option {
	return fx.Provide(
		service.NewUserService,
		service.NewAuthService,
		service.NewVaultService,
		service.NewCredentialService,
	)
}

func httpModule() fx.Option {
	return fx.Options(
		fx.Provide(http.NewServer),
		fx.Invoke(http.StartServer),
	)
}

func handlerModule() fx.Option {
	return fx.Provide(
		handler.NewAuthHandler,
		handler.NewVaultHandler,
		handler.NewCredentialHandler,
	)
}

func routerModule() fx.Option {
	return fx.Provide(
		router.RegisterRoutesFuncProvider,
		router.RegisterSwaggerRoutesFuncProvider,
	)
}

func docsModule() fx.Option {
	return fx.Provide(
		func() embed.FS {
			return embeddedFiles
		},
		docs.RegisterDocsRouterFuncProvider,
	)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	cmd := exec.Command("make", "swag")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Error running make swag: %v", err)
	}

	app := fx.New(
		configModule(),
		databaseModule(),
		repositoryModule(),
		serviceModule(),
		httpModule(),
		handlerModule(),
		routerModule(),
		docsModule(),
	)

	app.Run()
}
