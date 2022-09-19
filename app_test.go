package maryread

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	options := AppOptions{}
	app := New(options)
	assert.NotEmpty(t, app)
	assert.NotNil(t, app.router)
}

func TestNewAppWithCustomRouter(t *testing.T) {
	router := echo.New()
	options := AppOptions{Router: RouterOptions{Router: router}}
	app := New(options)
	assert.NotEmpty(t, app)
	assert.Equal(t, router, app.router)
}

func TestDefault(t *testing.T) {
	app := Default()
	assert.NotEmpty(t, app)
	assert.NotNil(t, app.router)
	assert.IsType(t, echo.New(), app.router)
}

func TestRouter(t *testing.T) {
	router := echo.New()
	options := AppOptions{Router: RouterOptions{Router: router}}
	app := New(options)
	assert.Equal(t, router, app.Router())

}
