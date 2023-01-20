# Mary Read

Provides a [echo router](https://echo.labstack.com/) wrapper with extra functionality like database initialization and util working middleware to be
shared between microservices.

Also helps to [fight global warming](https://en.wikipedia.org/wiki/Flying_Spaghetti_Monster#Pirates_and_global_warming) acting as [Mary Read](https://en.wikipedia.org/wiki/Mary_Read).

## App

### Default

Provides an app with default tools:

- A new echo router.
- The [RequestID echo middleware](https://echo.labstack.com/middleware/request-id/).
- A zero logger in the echo context.
- Prints requests logs.
- [Body Dump](https://echo.labstack.com/middleware/body-dump/) on header X-Bodydump not empty.
Usage:

```go
    import "github.com/orov-io/maryread
    // ...

    app := maryread.Default()
    // You can access to the new echo router by app.Router()
    // app.Router().Use()
    // app.Router().GET()...
    
```

### Custom App

Returns an app with tools configured via options:
Usage:

```go
    import "github.com/orov-io/maryread
    // ...

    e := echo.New()
    // e.GET()....
    options := maryread.AppOptions{
        Router: RouterOptions{
            Router: router
        }
    }
    
```

## Available Middleware

```go
    import "github.com/orov-io/maryread/middleware"
```

### Body Dump

Adds the default Bodydump echo middleware for request with the header *X-Bodydump* not empty.
Usage:

```go
e := echo.New()
e.Use(DefaultLogger(zerolog.DebugLevel))
e.Use(BodyDumpOnHeader())
```

It relies in the logger echo-midlleware (see below)
After this, all request with the *X-Bodydump* will use the zero logger functionality in this package to log both request & response bodies.

### Logger

Usage:

```go
e := echo.New()
e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey: []byte(testJWTSecret),
}))
e.Use(DefaultLogger(zerolog.DebugLevel))
```

Adds the zero logger to the echo context. It generates a child that:

- Adds the requestID if the requestID (the one in this package or the one prodived by echo) as a string param.
- Adds the userID if you uses the JWT echo middleware a a string params. At now, it spects the ID to be in the ID claim.
- Uses a pretty zerolog parser.
- Changes the log level for a single request if you send the *X-Loglevel* header set to a valid zerolog level value (trace, debug, info ...)

### RequestID

Usage:

```go
e := echo.New()
e.Use(middleware.RequestID())
// Inside a handler or middleware...
RequestID(c) # It returns the generated requestID
```

Returns the generated requestID if you use the echo requestID middleware

### Request Logger

Deprecated. Use echo.middleware.Logger() Instead.

## TODO list

[] Migrate current middleware to fit the echo provided middleware (a default initializer and an initializer with options... perhaps other initializer with default config overrides by env vars...)
[] Decouple auth middleware from firebase. Use the echo middleware with custom (parse) functions.
[] Use the echo.NewHTTPError to return errors that echo will understand, so return c.JSON(...) will not be used and code become clearest.
[] Add auth and sqlx middleware to readme. Remember to say to people must import desired driver in the file that they use to load the middleare, Examples:

- _ "github.com/mattn/go-sqlite3"
- _ "github.com/lib/pq"
- or another driver supported by sql (and goose if you want automigrations)


[] Add Must\<shortcut> to app, in order to panic if can obtain required object, as in dbx.
[] Add test to middleware shortcuts.
[] WARNING: If you want´t to use the datadog/sqlmock db mock in your test, please, deactivate the automigration feature or use the in memory sqlite3 driver (see the automigration test). We couldn't infer the Execs and transactions that goose made to the database in an Up command, so it will panic with no acction expecteds.
[] Test if the default injected logger by the default echo logger middleware generates childs or insulate loggers in each context. If true, se the inner echo logger provided by the logger middleware instead of zero logger, accepting another logger that fit the error interface. Do this will result in delete the requestID Middleware. If not, check https://github.com/labstack/echo/issues/2310 to come back to use the custom request logger by default. Also, force the auth middleware to set a header in the request (X-LoggedUserID) and use the header string template mechanism in the logger to expose it. 