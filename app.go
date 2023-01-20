package maryread

import (
	"firebase.google.com/go/auth"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/orov-io/maryread/middleware"
)

type App struct {
	router *echo.Echo
}

// AppOptions models the app tools.
type AppOptions struct {
	Router RouterOptions
}

// RouterOptions model the echo router options. If provided, app will use the
// expecified echo router.
type RouterOptions struct {
	Router    *echo.Echo
	Validator echo.Validator
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
	e.Use(echoMiddleware.Logger())
	e.Use(middleware.BodyDumpOnHeader())
	e.Validator = NewValidator()
	return e
}

// Router returns the inner echo router.
func (app *App) Router() *echo.Echo {
	return app.router
}

func routerFromOptions(options AppOptions) *echo.Echo {
	var e *echo.Echo
	if options.Router.Router != nil {
		e = options.Router.Router
	} else {
		e = echo.New()
	}

	if options.Router.Validator != nil {
		e.Validator = options.Router.Validator
	}

	return e
}

// RequestID is a shortcut to middleware.RequestID()
func RequestID(c echo.Context) string {
	return middleware.RequestID(c)
}

// GetIDToken is a shortcut to middleware.GetIDToken()
func GetIDToken(c echo.Context) (*auth.Token, error) {
	return middleware.GetIDToken(c)
}

// LoggedUserIs is a shortcut to middleware.LoggedUserIs()
func LoggedUserIs(c echo.Context, rol string) bool {
	return middleware.LoggedUserIs(c, rol)
}

// LoggedUserIsAny is a shortcut to middleware.LoggedUserIsAny()
func LoggedUserIsAny(c echo.Context, roles []string) bool {
	return middleware.LoggedUserIsAny(c, roles)
}

// GetDBX is a shortcut to middleware.GetDBX()
func GetDBX(c echo.Context) (*sqlx.DB, error) {
	return middleware.GetDBX(c)
}

// MustGetDBX is like GetDBX but panics on error
func MustGetDBX(c echo.Context) *sqlx.DB {
	dbx, err := GetDBX(c)
	if err != nil {
		panic(err)
	}
	return dbx
}
