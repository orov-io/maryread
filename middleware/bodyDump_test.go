package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

const bodyDumpTestPath = "/test"
const bodyDumpDefaultResponseKey = "message"
const bodyDumpDefaultResponse = "test"
const bodyDumpDefaultRequestMessage = "Truman"

const testBodyDumpLogLevel = "-"

type bodyDumpLog struct {
	Level        string
	RequestID    string
	RequestBody  string
	ResponseBody string
	Time         string
	Message      string
	Prefix       string
}

type defaultBodyDumpResponse struct {
	Message string `json:"message"`
}

func TestBodyDumpOnHeaderDefaultResponse(t *testing.T) {
	e := echo.New()
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))
	e.Use(BodyDumpOnHeader())

	handler, buffer := getTestBodyDumpHandler(t, nil)

	e.GET(bodyDumpTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, nil)
	req.Header[bodyDumpHeader] = []string{"true"}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var requestDump bodyDumpLog
	var responseDump bodyDumpLog
	rawBodyDump := buffer.Bytes()
	sliceBodyDump := strings.Split(string(rawBodyDump), "\n")
	assert.Len(t, sliceBodyDump, 3)

	json.Unmarshal([]byte(sliceBodyDump[0]), &requestDump)
	json.Unmarshal([]byte(sliceBodyDump[1]), &responseDump)

	assert.Equal(t, "X-Body-Dump/response", responseDump.Prefix)
	assert.Equal(t, "X-Body-Dump/request", requestDump.Prefix)

	assert.Empty(t, requestDump.Message)
	assert.NotEmpty(t, responseDump.Message)

	assert.Equal(t, testBodyDumpLogLevel, requestDump.Level)
	assert.Equal(t, testBodyDumpLogLevel, responseDump.Level)
}

func TestBodyDumpNoHeader(t *testing.T) {
	e := echo.New()
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))
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
	e.Use(ContextLogger(e.Logger, uint8(log.DEBUG)))
	e.Use(BodyDumpOnHeader())

	handler, buffer := getTestBodyDumpHandler(t, nil)

	e.GET(bodyDumpTestPath, handler)

	req := httptest.NewRequest(http.MethodGet, loggerTestPath, getDefaultRequestBody())
	req.Header[bodyDumpHeader] = []string{"true"}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var requestDump bodyDumpLog
	var responseDump bodyDumpLog
	rawBodyDump := buffer.Bytes()
	sliceBodyDump := strings.Split(string(rawBodyDump), "\n")
	assert.Len(t, sliceBodyDump, 3)

	json.Unmarshal([]byte(sliceBodyDump[0]), &requestDump)
	json.Unmarshal([]byte(sliceBodyDump[1]), &responseDump)

	assert.Equal(t, "X-Body-Dump/response", responseDump.Prefix)
	assert.Equal(t, "X-Body-Dump/request", requestDump.Prefix)

	assert.NotEmpty(t, requestDump.Message)
	assert.NotEmpty(t, responseDump.Message)

	assert.Equal(t, testBodyDumpLogLevel, requestDump.Level)
	assert.Equal(t, testBodyDumpLogLevel, responseDump.Level)
}

func getDefaultRequestBody() *bytes.Buffer {
	var buff bytes.Buffer
	json.NewEncoder(&buff).Encode(defaultBodyDumpResponse{Message: bodyDumpDefaultRequestMessage})
	return &buff
}

func getTestBodyDumpHandler(t *testing.T, resp *echo.Map) (echo.HandlerFunc, *bytes.Buffer) {
	buffer := new(bytes.Buffer)
	return func(c echo.Context) error {
		c.Logger().SetOutput(buffer)
		if resp == nil {
			resp = &echo.Map{bodyDumpDefaultResponseKey: bodyDumpDefaultResponse}
		}
		return c.JSON(http.StatusOK, resp)
	}, buffer
}
