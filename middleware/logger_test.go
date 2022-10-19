package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const loggerTestPath = "/test"
const loggerTestMsg = "test"

const expectedTestJWTID = "1234567890"
const testJWTSecret = "TrumanCapote"

type defaultLoggerResponse struct {
	Level     string
	RequestID string
	Time      string
	Message   string
	UserID    string
}

func TestDefaultLogger(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.DebugLevel))

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

	var log defaultLoggerResponse
	json.Unmarshal(data, &log)

	assert.Equal(t, "info", log.Level)
	assert.Empty(t, log.RequestID)
	assert.Equal(t, loggerTestMsg, log.Message)
	assert.Equal(t, userIDAnnonymousValue, log.UserID)
	responseTime, err := time.Parse(time.RFC3339, log.Time)
	assert.NoError(t, err)
	assert.LessOrEqual(t, time.Now().Unix(), responseTime.Unix())
}

func TestDefaultLoggerWithJWT(t *testing.T) {
	e := echo.New()
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(testJWTSecret),
		ContextKey: userContextField,
	}))
	e.Use(DefaultLogger(zerolog.DebugLevel))

	e.GET(loggerTestPath, getTestLoggerHandler(t))

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

	var log defaultLoggerResponse
	json.Unmarshal(data, &log)

	assert.Equal(t, expectedTestJWTID, log.UserID)
}

func TestDefaultLoggerChangingLevel(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.ErrorLevel))

	e.GET(loggerTestPath, getTestLoggerHandler(t))

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	rec := httptest.NewRecorder()
	req.Header.Set(logLevelHeader, zerolog.InfoLevel.String())

	e.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Unable to parse body: %v", err)
	}

	var log defaultLoggerResponse
	json.Unmarshal(data, &log)

	assert.Equal(t, "info", log.Level)
}

func getTestLoggerHandler(t *testing.T) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := GetLogger(c)
		buffer := new(bytes.Buffer)
		recorderLogger := logger.Output(buffer)
		recorderLogger.Info().Msg(loggerTestMsg)
		return c.String(http.StatusOK, buffer.String())
	}
}
