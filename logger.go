package maryRead

import (
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const loggerContextField = "zeroLogger"
const userContextField = "user"
const userIDAnnonymousValue = "annonymous"
const userIDLoggerContextualTag = "userID"
const idClaimKey = "ID"
const logLevelHeader = "X-Loglevel"

var defaultLevel = zerolog.DebugLevel

// Logger will inject in context a default zerolog logger with all
// available contextual info.
// Default logger is a pretty logger with info level and timestamp functionality.
func DefaultLogger(level zerolog.Level) echo.MiddlewareFunc {
	defaultLevel = level
	var logger = generateDefaultLogger()

	return injectLogger(logger)
}

func generateDefaultLogger() zerolog.Logger {
	var output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger().Level(defaultLevel)
}

func injectLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			var userID string
			switch user := c.Get(userContextField).(type) {
			case *jwt.Token:
				userID = user.Claims.(jwt.MapClaims)[idClaimKey].(string)

			default:
				userID = userIDAnnonymousValue
			}

			child := logger.With().
				Str("requestID", RequestID(c)).
				Str(userIDLoggerContextualTag, userID)

			SetLogger(c, child.Logger().Level(getLogLevel(c)))

			return next(c)
		}
	}
}

func getLogLevel(c echo.Context) zerolog.Level {
	logLevelHeader := c.Request().Header.Get(logLevelHeader)
	if logLevelHeader == "" {
		return defaultLevel
	}

	level, err := zerolog.ParseLevel(logLevelHeader)
	if err != nil {
		return defaultLevel
	}

	return level
}

// GetLogger returns the contextual logger from context
func GetLogger(c echo.Context) zerolog.Logger {
	switch logger := c.Get(loggerContextField).(type) {
	case zerolog.Logger:
		return logger

	default:
		return generateDefaultLogger()
	}

}

// SetLogger overrides the contextual logger in the echo context.
func SetLogger(c echo.Context, logger zerolog.Logger) {
	c.Set(loggerContextField, logger)
}
