package maryread

import "github.com/labstack/echo/v4"

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
		router: echo.New(),
	}
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
