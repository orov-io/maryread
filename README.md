# Mary Read
Provides a [echo router](https://echo.labstack.com/) wrapper with extra functionality like database initialization and util working middleware to be
shared between microservices.

Also helps to [fight global warming](https://en.wikipedia.org/wiki/Flying_Spaghetti_Monster#Pirates_and_global_warming) acting as [Mary Read](https://en.wikipedia.org/wiki/Mary_Read).

# App
## Default
Provides an app with default tools:
- A new echo router.
- The [RequestID echo middleware](https://echo.labstack.com/middleware/request-id/).
- A zero logger in the echo context.
- Prints requests logs.
- [Body Dump](https://echo.labstack.com/middleware/body-dump/) on header X-Bodydump not empty.
Usage:
```
    import "github.com/orov-io/maryread
    // ...

    app := maryread.Default()
    // You can access to the new echo router by app.Router()
    // app.Router().Use()
    // app.Router().GET()...
    
```

## Custom App
Returns an app with tools configured via options:
Usage:
```
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

# Available Middleware
```
    import "github.com/orov-io/maryread/middleware"
```
## Body Dump

Adds the default Bodydump echo middleware for request with the header *X-Bodydump* not empty.
Usage:
```
e := echo.New()
e.Use(DefaultLogger(zerolog.DebugLevel))
e.Use(BodyDumpOnHeader())
```

It relies in the logger echo-midlleware (see below)
After this, all request with the *X-Bodydump* will use the zero logger functionality in this package to log both request & response bodies.

## Logger
Usage:
```
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

## RequestID
Usage:
```
e := echo.New()
e.Use(middleware.RequestID())
// Inside a handler or middleware...
RequestID(c) # It returns the generated requestID
```
Returns the generated requestID if you use the echo requestID middleware

## Request Logger
Usage
```
e := echo.New()
e.Use(DefaultLogger(zerolog.DebugLevel))
e.Use(DefaultRequestZeroLoggerConfig())
```
It relies in the provided logger functionality, like the body dump middleware. Use the zerologger and the echo requestLogger middleware to log request info.