package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

const (
	loggerTestPath = "/test"
	loggerTestMsg  = "test"

	testLoggerDefaultLevel = log.DEBUG
	testLoggerINFOHeader   = "INFO"
	testLoggerPrefixHeader = "context"
	testLoggerFileHeader   = "logger_test.go"
)

type defaultLoggerResponse struct {
	Level     string
	RequestID string
	Time      string
	Message   string
	UserID    string
	Prefix    string
	File      string
	Line      string
}

func TestDefaultLogger(t *testing.T) {
	e := echo.New()
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))

	e.GET(loggerTestPath, getTestLoggerHandler(t))

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Unable to parse body: %v", err)
	}

	var logOutput defaultLoggerResponse
	json.Unmarshal(data, &logOutput)

	assert.Equal(t, testLoggerINFOHeader, logOutput.Level)
	assert.Empty(t, logOutput.RequestID)
	assert.Equal(t, loggerTestMsg, logOutput.Message)
	assert.Equal(t, testLoggerPrefixHeader, logOutput.Prefix)
	responseTime, err := time.Parse(time.RFC3339Nano, logOutput.Time)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, time.Now().UnixNano(), responseTime.UnixNano())
	_, err = strconv.Atoi(logOutput.Line)
	assert.NotEmpty(t, logOutput.Line)
	assert.NoError(t, err)
}

func TestDefaultLoggerWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	e.Use(authMiddleware.ParseJWT())
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))

	// TODO: This test is for the auth test. Remove the middleware for this handler.
	e.GET(loggerTestPath, getTestLoggerHandler(t), authMiddleware.LoggedUser())

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	rec := httptest.NewRecorder()
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("%v %v", testJWTHeaderPrefix, testJWT))

	e.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Unable to parse body: %v", err)
	}

	var logOutput defaultLoggerResponse
	json.Unmarshal(data, &logOutput)
	t.Logf("Tha log: %s", string(data))

	assert.Equal(t, mockAuthClientUID, logOutput.UserID)
}

func TestDefaultLoggerChangingLevel(t *testing.T) {
	e := echo.New()
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))

	e.GET(loggerTestPath, getTestLoggerHandler(t))

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	rec := httptest.NewRecorder()
	req.Header.Set(logLevelHeader, fmt.Sprint(log.WARN))

	e.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Unable to parse body: %v", err)
	}

	assert.Empty(t, data)
}

func getTestLoggerHandler(t *testing.T) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := c.Logger()
		buffer := new(bytes.Buffer)
		logger.SetOutput(buffer)
		logger.Info(loggerTestMsg)
		return c.String(http.StatusOK, buffer.String())
	}
}
