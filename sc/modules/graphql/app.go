package graphql

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/graphql-go/graphql"
	"github.com/spacecloud-io/space-cloud/managers/apis"
	"github.com/spacecloud-io/space-cloud/model"
	"github.com/spacecloud-io/space-cloud/modules/database"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(App{})
	apis.RegisterApp("graphql", 100)
}

// App manages all the database modules
type App struct {
	// Database app
	database *database.App

	// For internal usage
	logger    *zap.Logger
	dbSchemas map[string]model.DBSchemas

	// GraphQL schema
	schemas map[string]*graphql.Schema // Key: projectid
}

// CaddyModule returns the Caddy module information.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "graphql",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision sets up the file loader module.
func (a *App) Provision(ctx caddy.Context) error {
	a.logger = ctx.Logger(a)

	// Load the database app to retrieve all the parsed schemas
	dbApp, err := ctx.App("database")
	if err != nil {
		a.logger.Error("Unable to load database app", zap.Error(err), zap.String("app", "graphql"))
		return err
	}

	a.database = dbApp.(*database.App)
	a.dbSchemas = a.database.GetParsedSchemas()
	return nil
}

// Start begins the graphql app operations
func (a *App) Start() error {
	// Prepare schema for each project
	a.schemas = make(map[string]*graphql.Schema, len(a.dbSchemas))
	for project := range a.dbSchemas {
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: a.getQueryType(project),
		})
		if err != nil {
			a.logger.Error("Unable to prepare graphql schema", zap.Error(err), zap.String("project", project))
			return err
		}

		a.schemas[project] = &schema
	}
	return nil
}

// Stop ends the graphql app operations
func (a *App) Stop() error {
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.App         = (*App)(nil)
	_ apis.App          = (*App)(nil)
)