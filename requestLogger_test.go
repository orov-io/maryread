package maryRead

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type defaultRequestLoggerResponse struct {
	Level     string
	RequestID string
	Time      string
	Message   string
	Host      string
	URI       string `json:"URI"`
	Status    int
	Latency   string
}

const requestLoggerTestPath = "/test"
const requestLoggerTestMsg = "request"
const requestLoggerTestHost = "example.com"

func TestDefaultRequestZeroLoggerConfig(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.DebugLevel))
	e.Use(DefaultRequestZeroLoggerConfig())

	handler, buffer := getTestRequestLoggerHandler(t)

	e.GET(loggerTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, requestLoggerTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var log defaultRequestLoggerResponse
	json.Unmarshal(buffer.Bytes(), &log)

	assert.Equal(t, "info", log.Level)
	assert.Empty(t, log.RequestID)
	assert.Equal(t, requestLoggerTestHost, log.Host)
	assert.Equal(t, requestLoggerTestMsg, log.Message)
	assert.Equal(t, requestLoggerTestPath, log.URI)
	assert.Equal(t, http.StatusOK, log.Status)
	assert.NotEmpty(t, log.Latency)
	responseTime, err := time.Parse(time.RFC3339, log.Time)
	assert.NoError(t, err)
	assert.Greater(t, time.Now().UnixMilli(), responseTime.UnixMilli())
}

func getTestRequestLoggerHandler(t *testing.T) (echo.HandlerFunc, *bytes.Buffer) {
	buffer := new(bytes.Buffer)
	return func(c echo.Context) error {
		logger := GetLogger(c)
		recorderLogger := logger.Output(buffer)
		SetLogger(c, recorderLogger)
		return c.String(http.StatusOK, requestLoggerTestMsg)
	}, buffer
}
