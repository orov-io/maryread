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

const testJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjEyMzQ1Njc4OTAiLCJuYW1lIjoiVHJ1bWFuIENhcG90ZSIsInJvbGVzIjpbImFkbWluIl0sImV4cCI6MjI2MzEzNDQxM30.K_xzBV5s1omrF_7tzXaD0dvdhzUQGFd2VRVR7VmrjWE"
const expectedTestJWTID = "1234567890"
const testJWTSecret = "TrumanCapote"
const testJWTHeaderPrefix = "Bearer"

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

	t.Logf("tha raw body: %+v", string(data))

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

	t.Logf("tha raw body: %+v", string(data))

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
