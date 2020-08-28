package server

import (
	"github.com/gofiber/fiber"
	"github.com/trapck/gopref/model"
)

// Store serves as an interface for packages db operations
type Store interface {
	SavePkgs(p []model.Pkg) error
	SaveUsages(u []model.PkgImportUsage) error
	GetCombinedUsages() ([]model.CombinedUsage, error)
}

// App is an application to reflect go module packages info
type App struct {
	server *fiber.App
	store  Store
}

//Start starts an application
func (a *App) Start(port int) error {
	return a.server.Listen(port)
}

// NewApp initializes the new app instance
func NewApp(s Store) *App {
	app := &App{server: fiber.New(), store: s}
	app.server.Get("/packages/:user/:repo", app.HandlePackages)
	return app
}
