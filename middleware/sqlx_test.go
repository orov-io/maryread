package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const (
	sqlxTestPath                        = "/sqlxPath"
	sqlxTestHost                        = "localhost"
	sqlxTestPort                        = "5453"
	sqlxTestUser                        = "truman"
	sqlxTestPassword                    = "capote"
	sqlxTestDBName                      = "trumanCapote"
	sqlxTestSSLMode                     = "disable"
	sqlxTestExpectedPSQLInfo            = "host=localhost port=5453 user=truman password=capote dbname=trumanCapote sslmode=disable"
	sqlxTestAbsoluteMigrationPathEnvKey = "SQLX_MIGRATION_PATH"
)

func TestSQLXWithConfigWithDBAndDriver(t *testing.T) {
	db, _, _ := sqlmock.New()
	config := SQLXConfig{
		DB:     db,
		Driver: defaultSQLDriver,
	}
	var res *http.Response

	assert.NotPanics(t, func() {
		_, _, res = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestSQLXWithConfigWithDBNoDriver(t *testing.T) {
	db, _, _ := sqlmock.New()
	config := SQLXConfig{
		DB: db,
	}

	assert.Panics(t, func() {
		_, _, _ = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})
}

func TestSQLXWithConfigNoDataSource(t *testing.T) {
	config := SQLXConfig{
		Driver: "someDriver",
	}
	assert.Panics(t, func() {
		_, _, _ = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})
}

func TestSQLXAutomigrateNoPath(t *testing.T) {
	db, _, _ := sqlmock.New()
	config := SQLXConfig{
		DB:          db,
		Driver:      defaultSQLDriver,
		AutoMigrate: true,
	}

	assert.Panics(t, func() {
		_, _, _ = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})

}

func TestSQLXAutomigrateSuccess(t *testing.T) {
	config := SQLXConfig{
		Driver:         "sqlite3",
		DataSourceName: ":memory:",
		AutoMigrate:    true,
		MigrationPath:  fmt.Sprintf("%s/success", os.Getenv(sqlxTestAbsoluteMigrationPathEnvKey)),
	}

	assert.NotPanics(t, func() {
		_, _, _ = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})
}

func TestSQLXAutomigrateFail(t *testing.T) {
	config := SQLXConfig{
		Driver:         "sqlite3",
		DataSourceName: ":memory:",
		AutoMigrate:    true,
		MigrationPath:  fmt.Sprintf("%s/fail", os.Getenv(sqlxTestAbsoluteMigrationPathEnvKey)),
	}

	assert.Panics(t, func() {
		_, _, _ = sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config)
	})
}

func TestParseSQLXEnvVars(t *testing.T) {
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxHostEnvKey, sqlxTestHost)
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxPortEnvKey, sqlxTestPort)
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxUserEnvKey, sqlxTestUser)
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxPasswordEnvKey, sqlxTestPassword)
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxDBNameEnvKey, sqlxTestDBName)
	assert.Panics(t, sqlxTestTryParseEnvVars)

	os.Setenv(sqlxSSLModeEnvKey, sqlxTestSSLMode)
	host, port, user, password, dbName, sslMode := parseSQLXEnvVars()

	assert.Equal(t, sqlxTestHost, host)
	assert.Equal(t, sqlxTestPort, port)
	assert.Equal(t, sqlxTestUser, user)
	assert.Equal(t, sqlxTestPassword, password)
	assert.Equal(t, sqlxTestDBName, dbName)
	assert.Equal(t, sqlxTestSSLMode, sslMode)
}

func TestGeneratePSQLInfo(t *testing.T) {
	os.Setenv(sqlxHostEnvKey, sqlxTestHost)
	os.Setenv(sqlxPortEnvKey, sqlxTestPort)
	os.Setenv(sqlxUserEnvKey, sqlxTestUser)
	os.Setenv(sqlxPasswordEnvKey, sqlxTestPassword)
	os.Setenv(sqlxDBNameEnvKey, sqlxTestDBName)
	os.Setenv(sqlxSSLModeEnvKey, sqlxTestSSLMode)

	psqlInfo := generatePSQLInfo()
	assert.Equal(t, sqlxTestExpectedPSQLInfo, psqlInfo)
}

func sqlxTestTryParseEnvVars() {
	parseSQLXEnvVars()
}

func TestGetDBNoMiddleware(t *testing.T) {
	e := echo.New()
	e.GET(sqlxTestPath, sqlxTestHandler)
	req := httptest.NewRequest(http.MethodGet, sqlxTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	res := rec.Result()
	res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func sqlxTestGetRouterAndRequestWithMiddleareAndTestHandlerClosingBody(config SQLXConfig) (
	*echo.Echo, *http.Request, *http.Response,
) {

	e := echo.New()
	m := NewSQLX()
	e.Use(m.WithConfig(config))
	e.GET(sqlxTestPath, sqlxTestHandler)
	req := httptest.NewRequest(http.MethodGet, sqlxTestPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	res := rec.Result()
	res.Body.Close()

	return e, req, res

}

func sqlxTestHandler(c echo.Context) error {
	_, err := GetDBX(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
