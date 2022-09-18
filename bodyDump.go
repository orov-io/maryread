package middleware

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const requestBodyKey = "requestBody"
const responseBodyKey = "responseBody"
const bodyDumpMessage = "Body Dump"
const bodyDumpHeader = "X-Bodydump"

func BodyDumpOnHeader() echo.MiddlewareFunc {
	return middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{

		Handler: bodyDumpHanler,

		Skipper: func(c echo.Context) bool {
			header := c.Request().Header[bodyDumpHeader]
			return len(header) == 0
		},
	})
}

func bodyDumpHanler(c echo.Context, reqBody, resBody []byte) {
	logger := GetLogger(c)

	logger.Debug().
		Str(requestBodyKey, cleanStringlifyBody(reqBody)).
		Str(responseBodyKey, cleanStringlifyBody(resBody)).
		Msg(bodyDumpMessage)
}

func cleanStringlifyBody(body []byte) string {
	var JSON map[string]interface{}
	err := json.Unmarshal(body, &JSON)
	if err != nil {
		return string(body)
	}

	cleanBody, err := json.Marshal(JSON)
	if err != nil {
		return string(body)
	}

	return string(cleanBody)
}
