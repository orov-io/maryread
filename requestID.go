package middleware

import "github.com/labstack/echo/v4"

func RequestID(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}
