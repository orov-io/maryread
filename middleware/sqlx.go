package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
)

type SQLXConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper em.Skipper

	// BeforeFunc defines a function which is executed just before the middleware.
	BeforeFunc em.BeforeFunc

	// DB is the inner db to be used. If is not provided, a new db will be open based either
	// in the Driver and DataSourceName attribute or in the default env vars.
	DB *sql.DB

	// Driver defines the driver to be used when a new database is tried to be opened.
	// If DB is provided, the middleware will never tried to stablish this connection.
	Driver string

	// DataSorceName is used in conjuntion with the Driver attribute when a new database is tried to be opened.
	// If DB is provided, the middleware will never tried to stablish this connection.
	DataSourceName string
}

const (
	defaultSQLDriver = "postgres"
	sqlxDBContextKey = "dbx"
)

var (
	sqlxInnerInstance *sqlx.DB
	sqlxIsInitialized bool
	DefaultSQLXConfig = SQLXConfig{
		Skipper:        em.DefaultSkipper,
		DB:             nil,
		Driver:         "",
		DataSourceName: "",
	}
)

func SQLX() echo.MiddlewareFunc {
	config := DefaultSQLXConfig
	config.Driver = defaultSQLDriver
	config.DataSourceName = generatePSQLInfo()
	return SQLXWithConfig(config)
}

func SQLXWithConfig(config SQLXConfig) echo.MiddlewareFunc {
	config = mixSQLXConfigDefault(config)
	sqlxInnerInstance = initDB(config)
	sqlxIsInitialized = true
	return sqlxHandlerFunc(config)
}

func initDB(config SQLXConfig) *sqlx.DB {
	if sqlxIsInitialized {
		panic("SQLX middleware already initialized!")
	}

	var dbx *sqlx.DB
	if config.DB != nil {
		dbx = sqlx.NewDb(config.DB, config.Driver)
	} else {
		dbx = sqlx.MustOpen(config.Driver, config.DataSourceName)
	}

	err := dbx.Ping()
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database with provided config due to error %v", err))
	}

	return dbx
}

func sqlxHandlerFunc(config SQLXConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil {
				return next(c)
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			c.Set(sqlxDBContextKey, sqlxInnerInstance)
			return next(c)
		}
	}
}

func mixSQLXConfigDefault(config SQLXConfig) SQLXConfig {
	if config.Skipper == nil {
		config.Skipper = DefaultSQLXConfig.Skipper
	}

	if config.DB != nil {
		if config.Driver == "" {
			panic("To use a existing sql.DB database you must provide the driver to be used in the Driver attribute.")
		}
		return config
	}

	//
	if (config.Driver == "" || config.Driver == defaultSQLDriver) && config.DataSourceName == "" {
		config.Driver = defaultSQLDriver
		config.DataSourceName = generatePSQLInfo()
	}

	if config.DataSourceName == "" {
		panic("Please, specify either the pair DataSourceConnection and Driver or suply a valid DB.")
	}

	return config
}

func generatePSQLInfo() string {
	host, port, user, password, dbname := parseSQLXEnvVars()
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

const (
	sqlxHostEnvKey     = "POSTGRES_HOST"
	sqlxPortEnvKey     = "POSTGRES_PORT"
	sqlxUserEnvKey     = "POSTGRES_USER"
	sqlxPasswordEnvKey = "POSTGRES_PASSWORD"
	sqlxDBNameEnvKey   = "POSTGRES_DBNAME"
)

func parseSQLXEnvVars() (host, port, user, password, dbName string) {
	var ok bool
	host, ok = os.LookupEnv(sqlxHostEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxHostEnvKey)
	}

	port, ok = os.LookupEnv(sqlxPortEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxPortEnvKey)
	}

	user, ok = os.LookupEnv(sqlxUserEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxUserEnvKey)
	}

	password, ok = os.LookupEnv(sqlxPasswordEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxPasswordEnvKey)
	}

	dbName, ok = os.LookupEnv(sqlxDBNameEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxDBNameEnvKey)
	}

	return
}

func panicBySQLXEnv(key string) {
	panic(fmt.Sprintf("Using default SQLX middleware. Please, specify %s in the provided env vars", key))
}

var ErrDBXMissing = echo.NewHTTPError(http.StatusInternalServerError, "unable to obtain the dbx in context. Please, initiate the sqlx middleware first")

func GetDBX(c echo.Context) (*sqlx.DB, error) {
	switch dbx := c.Get(sqlxDBContextKey).(type) {
	case *sqlx.DB:
		if dbx == nil {
			return dbx, ErrDBXMissing
		}
		return dbx, nil
	}
	return nil, ErrDBXMissing

}
