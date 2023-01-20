package maryread

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const testDefaultMiddlewarePath = "/test"

var testDefaultMiddlewareslogger zerolog.Logger
var testDefaultMiddlewaresRequestID string

func TestNewApp(t *testing.T) {
	options := AppOptions{}
	app := New(options)
	assert.NotEmpty(t, app)
	assert.NotNil(t, app.router)
}

func TestNewAppWithCustomRouter(t *testing.T) {
	router := echo.New()
	options := AppOptions{Router: RouterOptions{Router: router, Validator: NewValidatorRawError()}}
	app := New(options)
	assert.NotEmpty(t, app)
	assert.Equal(t, router, app.router)
	assert.NotNil(t, app.Router().Validator)
	assert.IsType(t, &ValidatorRawError{}, app.Router().Validator)
}

func TestRouter(t *testing.T) {
	router := echo.New()
	options := AppOptions{Router: RouterOptions{Router: router}}
	app := New(options)
	assert.Equal(t, router, app.Router())
	assert.Nil(t, app.Router().Validator)
}

func TestDefault(t *testing.T) {
	app := Default()
	assert.NotEmpty(t, app)
	assert.NotNil(t, app.router)
	assert.IsType(t, echo.New(), app.router)
	assert.NotNil(t, app.Router().Validator)
	assert.IsType(t, &Validator{}, app.Router().Validator)
}

func TestDefaultMiddlewares(t *testing.T) {
	app := Default()
	app.Router().GET(testDefaultMiddlewarePath, getTestDefaultMiddlewaresHandler(t))

	req := httptest.NewRequest(http.MethodGet, testDefaultMiddlewarePath, nil)
	rec := httptest.NewRecorder()

	app.Router().ServeHTTP(rec, req)

	// TODO: Test the logger.

	// RequestID
	assert.NotEmpty(t, testDefaultMiddlewaresRequestID)
}

func getTestDefaultMiddlewaresHandler(t *testing.T) echo.HandlerFunc {
	return func(c echo.Context) error {
		testDefaultMiddlewaresRequestID = RequestID(c)
		return c.String(http.StatusOK, "test")
	}
}
