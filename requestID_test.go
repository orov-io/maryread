package maryRead

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

const requestIDTestPath = "/test"

func TestRequestID(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())

	e.GET(requestIDTestPath, getTestRequestIDHandler(t))

	req := httptest.NewRequest(http.MethodGet, requestIDTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
}

func getTestRequestIDHandler(t *testing.T) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := RequestID(c)
		assert.Equal(t, requestID, c.Response().Header().Get(echo.HeaderXRequestID))
		return c.String(http.StatusOK, "test")
	}
}
