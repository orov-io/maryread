package middleware

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const bodyDumpHeader = "X-Body-Dump"

func BodyDumpOnHeader() echo.MiddlewareFunc {
	return middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{

		Handler: bodyDumpHandler,

		Skipper: func(c echo.Context) bool {
			header := c.Request().Header[bodyDumpHeader]
			return len(header) == 0
		},
	})
}

func bodyDumpHandler(c echo.Context, reqBody, resBody []byte) {
	printBody(c, reqBody, "request")
	printBody(c, resBody, "response")
}

func printBody(c echo.Context, body []byte, prefix string) {
	oldPrefix := c.Logger().Prefix()

	c.Logger().SetPrefix(fmt.Sprintf("%s/%s", bodyDumpHeader, prefix))
	c.Logger().Printf(cleanStringlifyBody(body))

	c.Logger().SetPrefix(oldPrefix)
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
