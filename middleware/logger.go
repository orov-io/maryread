package middleware

import (
	"fmt"
	"io"
	"strconv"

	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	logLevelHeader             = "X-Log-Level"
	contextLoggerPanicHeader   = "[Context Logger]"
	defaultContextLoggerHeader = `{"time":"${time_rfc3339_nano},"requestID":"${header:X-Request-ID}","level":"${level}","userID":"${header:X-Logged-User-Id}","prefix":"${prefix}","file":"${short_file}","line":"${line}"}"`
)

type (
	ContextLoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper em.Skipper

		// BeforeFunc defines a function which is executed just before the middleware.
		BeforeFunc em.BeforeFunc

		// Logger defines the interface of the logger to inject to the context
		Logger echo.Logger

		// Level defines the log level. It uses the github.com/labstack/gommon/log levels.
		// Let it empty to preserve the one sets in provided Logger.
		Level uint8

		// Output represent the output stream to write the log.
		// Let it nil to preserve the one sets in provided Logger.
		Output io.Writer

		// Header defines the template to use to print logs. See github.com/labstack/gommon/log.
		// Set it can cause lost the id attribute that logs the request ID.
		// Let it empty to preserve the one sets in provided Logger.
		Header string

		// Header defines the prefix to print in logs, if defined in Header. See github.com/labstack/gommon/log.
		// Defaults set to "context".
		// Let it empty to preserve the one sets in provided Logger.
		Prefix string
	}
)

var ContextLoggerDefaultConfig = ContextLoggerConfig{
	Skipper: em.DefaultSkipper,
	Prefix:  "context",
}

func ContextLogger(logger echo.Logger, level uint8) echo.MiddlewareFunc {
	config := ContextLoggerDefaultConfig
	config.Logger = logger
	config.Level = level
	config.Header = defaultContextLoggerHeader
	return ContextLoggerWithConfig(config)
}

func ContextLoggerWithConfig(config ContextLoggerConfig) echo.MiddlewareFunc {
	mixContextLoggerDefaultConfig(&config)
	mustSetLoggerLevel(config.Logger, config.Level)
	if config.Output != nil {
		config.Logger.SetOutput(config.Output)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetLogger(log.New(config.Prefix))
			c.Logger().SetOutput(config.Output)
			setLoggerHeader(c, config)
			c.Logger().SetLevel(getLogLevelFromContext(c, config.Level))

			return next(c)
		}
	}
}

func mixContextLoggerDefaultConfig(config *ContextLoggerConfig) {
	if config.Logger == nil {
		panic(fmt.Sprintf("%s Please, provide a not nil logger in config", contextLoggerPanicHeader))
	}

	if config.Skipper == nil {
		config.Skipper = ContextLoggerDefaultConfig.Skipper
	}

	if config.Level == 0 {
		config.Level = uint8(config.Logger.Level())
	}

	if config.Prefix == "" {
		config.Prefix = config.Logger.Prefix()
	}

	if config.Output == nil {
		config.Output = config.Logger.Output()
	}
}

func setLoggerHeader(c echo.Context, config ContextLoggerConfig) {
	if config.Header == "" {
		return
	}

	c.Logger().SetHeader(config.Header)
}

func getLogLevelFromContext(c echo.Context, fallbackLevel uint8) log.Lvl {
	logLevelHeader := c.Request().Header.Get(logLevelHeader)
	if logLevelHeader == "" {
		return log.Lvl(fallbackLevel)
	}

	u64, err := strconv.ParseUint(logLevelHeader, 10, 32)
	if err != nil || u64 > 7 {
		c.Logger().Errorf("Invalid log level in header %s. Please, provide a int between 1 and 7", logLevelHeader)
		return c.Logger().Level()
	}

	return log.Lvl(uint8(u64))
}

func isValidLogLevel(level uint8) bool {
	return level > 0 && level < 8
}

func mustSetLoggerLevel(logger echo.Logger, level uint8) {
	if !isValidLogLevel(level) {
		panic(fmt.Sprintf("%s 0<LogLevel<8, %d provided", contextLoggerPanicHeader, level))
	}

	logger.SetLevel(log.Lvl(level))
}
