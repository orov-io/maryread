package maryRead

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const bodyDumpTestPath = "/test"
const bodyDumpDefaultResponseKey = "message"
const bodyDumpDefaultResponse = "test"
const bodyDumpDefautlRequestMessage = "Truman"

const bodyDumpLogLevel = "debug"

type bodyDumpLog struct {
	Level        string
	RequestID    string
	RequestBody  string
	ResponseBody string
	Time         string
	Message      string
}

type defaultBodyDumpResponse struct {
	Message string `json:"message"`
}

func TestBodyDumpOnHeaderDefaultResponse(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.DebugLevel))
	e.Use(BodyDumpOnHeader())

	handler, buffer := getTestBodyDumpHandler(t, nil)

	e.GET(bodyDumpTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	req.Header[bodyDumpHeader] = []string{"true"}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var log bodyDumpLog
	json.Unmarshal(buffer.Bytes(), &log)

	var responseBody defaultBodyDumpResponse
	json.Unmarshal([]byte(log.ResponseBody), &responseBody)

	assert.Equal(t, bodyDumpLogLevel, log.Level)
	assert.Empty(t, log.RequestID)
	assert.Empty(t, log.RequestBody)
	assert.NotEmpty(t, log.ResponseBody)
	assert.Equal(t, bodyDumpDefaultResponse, responseBody.Message)
}

func TestBodyDumpNoHeader(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.DebugLevel))
	e.Use(BodyDumpOnHeader())

	handler, buffer := getTestBodyDumpHandler(t, nil)

	e.GET(bodyDumpTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var log bodyDumpLog
	json.Unmarshal(buffer.Bytes(), &log)

	var responseBody defaultBodyDumpResponse
	json.Unmarshal([]byte(log.ResponseBody), &responseBody)

	assert.Empty(t, log.RequestBody)
	assert.Empty(t, log.ResponseBody)
}

func TestBodyDumpOnHeaderDefaultResponseWhitRequestBody(t *testing.T) {
	e := echo.New()
	e.Use(DefaultLogger(zerolog.DebugLevel))
	e.Use(BodyDumpOnHeader())

	handler, buffer := getTestBodyDumpHandler(t, nil)

	e.GET(bodyDumpTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, getDefaultRequestBody())
	req.Header[bodyDumpHeader] = []string{"true"}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var log bodyDumpLog
	json.Unmarshal(buffer.Bytes(), &log)

	var responseBody defaultBodyDumpResponse
	json.Unmarshal([]byte(log.ResponseBody), &responseBody)

	var requestBody defaultBodyDumpResponse
	json.Unmarshal([]byte(log.RequestBody), &requestBody)

	assert.Equal(t, bodyDumpLogLevel, log.Level)
	assert.Empty(t, log.RequestID)
	assert.NotEmpty(t, log.RequestBody)
	assert.Equal(t, bodyDumpDefautlRequestMessage, requestBody.Message)
	assert.NotEmpty(t, log.ResponseBody)
	assert.Equal(t, bodyDumpDefaultResponse, responseBody.Message)
}

func getDefaultRequestBody() *bytes.Buffer {
	var buff bytes.Buffer
	json.NewEncoder(&buff).Encode(defaultBodyDumpResponse{Message: bodyDumpDefautlRequestMessage})
	return &buff
}

func getTestBodyDumpHandler(t *testing.T, resp *echo.Map) (echo.HandlerFunc, *bytes.Buffer) {
	buffer := new(bytes.Buffer)
	return func(c echo.Context) error {
		logger := GetLogger(c)
		recorderLogger := logger.Output(buffer)
		SetLogger(c, recorderLogger)
		if resp == nil {
			resp = &echo.Map{bodyDumpDefaultResponseKey: bodyDumpDefaultResponse}
		}
		return c.JSON(http.StatusOK, resp)
	}, buffer
}
