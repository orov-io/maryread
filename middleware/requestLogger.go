package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func DefaultRequestZeroLoggerConfig() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogLatency: true,
		LogHost:    true,
		LogMethod:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger := GetLogger(c)
			logger.Info().
				Str("host", v.Host).
				Str("URI", v.URI).
				Str("method", v.Method).
				Int("status", v.Status).
				Str("latency", v.Latency.String()).
				Msg("request")

			return nil
		},
	})
}
