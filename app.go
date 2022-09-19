package maryread

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/orov-io/maryread/middleware"
	"github.com/rs/zerolog"
)

type App struct {
	router *echo.Echo
}

// NewApp generates a new app with tools expecified in provided options.
func New(options AppOptions) *App {
	return &App{
		router: routerFromOptions(options),
	}
}

// Returns a new app with default values.
func Default() *App {
	return &App{
		router: getEchoWithDefaultMiddleware(),
	}
}

func getEchoWithDefaultMiddleware() *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.DefaultLogger(zerolog.DebugLevel))
	e.Use(middleware.DefaultRequestZeroLoggerConfig())
	e.Use(middleware.BodyDumpOnHeader())
	return e
}

// AppOptions models the app tools.
type AppOptions struct {
	Router RouterOptions
}

// RouterOptions model the echo router options. If provided, app will use the
// expecified echo router.
type RouterOptions struct {
	Router *echo.Echo
}

// Router returns the inner echo router.
func (app *App) Router() *echo.Echo {
	return app.router
}

func routerFromOptions(options AppOptions) *echo.Echo {
	if options.Router.Router != nil {
		return options.Router.Router
	}

	return echo.New()
}

// GetLogger is a shortcut to middleware.GetLogger()
func GetLogger(c echo.Context) zerolog.Logger {
	return middleware.GetLogger(c)
}

// RequestID is a shortcut to middleware.RequestID()
func RequestID(c echo.Context) string {
	return middleware.RequestID(c)
}
