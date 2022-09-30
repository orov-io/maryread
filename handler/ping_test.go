package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewPing(t *testing.T) {
	ping := NewPingHandler()
	assert.NotNil(t, ping)
}

func TestGetPingHandler(t *testing.T) {
	e := getEchoRouterWithPingHandlers()

	req := httptest.NewRequest(http.MethodGet, pingPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Unable to parse body: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, pongResponse, string(data))
}

func getEchoRouterWithPingHandlers() *echo.Echo {
	e := echo.New()
	ping := NewPingHandler()
	ping.AddHandlers(e)
	return e
}
