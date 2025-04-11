package docs

import (
  "bytes"
  "embed"
  "fmt"
  "html/template"
  "math/rand"
  "path"

  "github.com/gofiber/fiber/v3"
)

type Provider struct {
  URL   string
  Name  string
  Theme string
}

func randomTheme() string {
  themes := []string{"bluePlanet", "deepSpace", "kepler"}
  return themes[rand.Intn(len(themes))]
}
func RegisterDocsRouter(fs embed.FS) func(app *fiber.App) {
  return func(app *fiber.App) {
    app.Get("/public/openapi.json", func(ctx fiber.Ctx) error {
      data, err := fs.ReadFile("docs/swagger.json")
      if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Fail to read swagger.json: "+err.Error())
      }
      ctx.Type("json")
      return ctx.Send(data)
    })

    app.Get("/docs", func(ctx fiber.Ctx) error {
      name := ctx.Query("name", "scalar")

      filename := path.Join("pkg/docs", fmt.Sprintf("%s.html", name))
      data, err := fs.ReadFile(filename)
      if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Fail to read template: "+err.Error())
      }

      tmpl, err := template.New(name).Parse(string(data))
      if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Fail to parse template: "+err.Error())
      }

      provider := &Provider{
        URL:   "/public/openapi.json",
        Name:  name,
        Theme: randomTheme(),
      }

      var bufferHTML bytes.Buffer
      if err := tmpl.Execute(&bufferHTML, provider); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Fail to execute template: "+err.Error())
      }

      return ctx.Type("html").SendString(bufferHTML.String())
    })
  }
}

type RegisterDocsRouterFunc func(app *fiber.App)

// RegisterDocsRouterFuncProvider provider for Fx
func RegisterDocsRouterFuncProvider(fs embed.FS) RegisterDocsRouterFunc {
  return RegisterDocsRouter(fs)
}
