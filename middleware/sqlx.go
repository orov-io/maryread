package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	goose "github.com/pressly/goose/v3"
)

type (
	SQLXConfig struct {
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

		// AutoMigrate defines if migrations will apply on middleware initialization. If true,
		// you must provide a MigrationPath folder with the migrations files.
		// pressly/goose (https://github.com/pressly/goose) is used to execute the migrations.
		// The default SQLX middleware set this attribute to true
		AutoMigrate bool

		// MigrationPath defines the path to migrations files. Is used on middleware initialization
		// If AutoMigrate is set to true. Also used in calls to SQLXApplyMigrations.
		// pressly/goose (https://github.com/pressly/goose) is used to execute the migrations.
		// The default SQLX middleware set this attribute to ./migration
		MigrationPath string
	}
	SQLX struct {
		config      SQLXConfig
		dbx         *sqlx.DB
		initialized bool
	}
)

const (
	defaultSQLDriver = "postgres"
	sqlxDBContextKey = "dbx"
)

var (
	DefaultSQLXConfig = SQLXConfig{
		Skipper:        em.DefaultSkipper,
		DB:             nil,
		Driver:         "",
		DataSourceName: "",
		AutoMigrate:    false,
		MigrationPath:  "",
	}
)

func NewSQLX() *SQLX {
	return &SQLX{
		config:      SQLXConfig{},
		dbx:         nil,
		initialized: false,
	}
}

func (m *SQLX) Default() echo.MiddlewareFunc {
	config := DefaultSQLXConfig
	config.Driver = defaultSQLDriver
	config.DataSourceName = generatePSQLInfo()
	config.AutoMigrate = true
	config.MigrationPath = "./migration"
	return m.WithConfig(config)
}

func (m *SQLX) WithConfig(config SQLXConfig) echo.MiddlewareFunc {
	m.config = config
	m.mixSQLXConfigDefault()
	m.dbx = m.initDB()
	m.initialized = true
	m.autoApplyMigrations()
	return m.sqlxHandlerFunc()
}

func (m *SQLX) initDB() *sqlx.DB {
	if m.initialized {
		panic("SQLX middleware already initialized!")
	}

	dbx, err := m.mustOpenDB()
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database with provided config due to error %v", err))
	}

	return dbx
}

func (m *SQLX) mustOpenDB() (*sqlx.DB, error) {
	ret := retrier.New(retrier.ExponentialBackoff(5, 1*time.Second), retrier.DefaultClassifier{})

	var dbx *sqlx.DB
	err := ret.Run(
		func() error {
			if m.config.DB != nil {
				dbx = sqlx.NewDb(m.config.DB, m.config.Driver)
			} else {
				dbx = sqlx.MustOpen(m.config.Driver, m.config.DataSourceName)
			}

			return dbx.Ping()
		})

	return dbx, err
}

func (m *SQLX) autoApplyMigrations() {
	if !m.config.AutoMigrate {
		return
	}

	goose.SetDialect(m.dbx.DriverName())
	err := m.applyMigration(m.config.MigrationPath)
	if err != nil {
		panic(fmt.Sprintf("Unable to apply SQLX migrations due to error: %v", err))
	}
}

func (m *SQLX) applyMigration(migrationPath string) error {
	return goose.Up(m.dbx.DB, migrationPath)
}

func (m *SQLX) sqlxHandlerFunc() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper(c) {
				return next(c)
			}

			if m.config.BeforeFunc != nil {
				m.config.BeforeFunc(c)
			}

			c.Set(sqlxDBContextKey, m.dbx)
			return next(c)
		}
	}
}

func (m *SQLX) mixSQLXConfigDefault() {
	if m.config.Skipper == nil {
		m.config.Skipper = DefaultSQLXConfig.Skipper
	}

	if m.config.DB != nil {
		m.panicNoDriver()
	} else {
		m.adjustDriverAndDataSourceConfigOrPanic()
	}

	if m.config.AutoMigrate && m.config.MigrationPath == "" {
		panic("To enable SQLX automigrations you must specify the migrations path in config.MigrationPath")
	}
}

func (m *SQLX) panicNoDriver() {
	if m.config.Driver == "" {
		panic("To use a existing sql.DB database you must provide the driver to be used in the Driver attribute.")
	}
}

func (m *SQLX) adjustDriverAndDataSourceConfigOrPanic() {
	if (m.config.Driver == "" || m.config.Driver == defaultSQLDriver) && m.config.DataSourceName == "" {
		m.config.Driver = defaultSQLDriver
		m.config.DataSourceName = generatePSQLInfo()
	}

	if m.config.DataSourceName == "" {
		panic("Please, specify either the pair DataSourceConnection and Driver or suply a valid DB.")
	}
}

func generatePSQLInfo() string {
	host, port, user, password, dbname, sslMode := parseSQLXEnvVars()
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslMode)
}

const (
	sqlxHostEnvKey     = "POSTGRES_HOST"
	sqlxPortEnvKey     = "POSTGRES_PORT"
	sqlxUserEnvKey     = "POSTGRES_USER"
	sqlxPasswordEnvKey = "POSTGRES_PASSWORD"
	sqlxDBNameEnvKey   = "POSTGRES_DBNAME"
	sqlxSSLModeEnvKey  = "POSTGRES_SSLMODE"
)

func parseSQLXEnvVars() (host, port, user, password, dbName, sslMode string) {
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

	sslMode, ok = os.LookupEnv(sqlxSSLModeEnvKey)
	if !ok {
		panicBySQLXEnv(sqlxSSLModeEnvKey)
	}

	return
}

func panicBySQLXEnv(key string) {
	panic(fmt.Sprintf("Using default SQLX middleware. Please, specify %s in the provided env vars", key))
}

var ErrDBXMissing = echo.NewHTTPError(http.StatusInternalServerError, "unable to obtain the dbx in context. Please, initiate the sqlx middleware first")

func GetDBX(c echo.Context) (*sqlx.DB, error) {
	dbx, ok := c.Get(sqlxDBContextKey).(*sqlx.DB)
	if !ok || dbx == nil {
		return dbx, ErrDBXMissing
	}
	return dbx, nil

}
