package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/labstack/echo/v4"
)

const (
	bearerPrefix = "Bearer "
)

var app *firebase.App
var client *auth.Client
var firebaseInitialized = false

type AuthMiddleware struct {
	authClient AuthClient
	ctx        context.Context
}

type AuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// DefaultAuthMiddleware returns the auth middleware with a firebase auth client initialized from env vars.
func DefaultAuthMiddleware() *AuthMiddleware {
	if !firebaseInitialized {
		initFirebase()
	}
	return NewAuthMiddleware(context.Background(), client)
}

// NewAuthMiddleware return a new auth middleware with the desired authClient attached.
func NewAuthMiddleware(ctx context.Context, authClient AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		ctx:        ctx,
	}
}

// AllowAnonymous will let pass all petitions trying to find a JWT in headers and log
// in the user if the JWT is found.
func (a *AuthMiddleware) AllowAnonymous() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, _ = a.login(c)
			return next(c)
		}
	}
}

// ParseJWT tries to find the JWT in auth header and parse it, letting the parsed jwt token
// in both in context and in the X-Logged-User-ID headers.
// Its a shortcut of AllowAnonymous method, with a better name to apply as a top level middleware.
// Useful to have user information in middlewares before the "per path" middlewares.
func (a *AuthMiddleware) ParseJWT() echo.MiddlewareFunc {
	return a.AllowAnonymous()
}

// LoggedUsers search for a valid JWT and logs in the founded user. If no JWT is found,
// it returns a 401 unauthorized standard error, stopping the request.
func (a *AuthMiddleware) LoggedUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := a.login(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": err.Error(),
				})
			}
			return next(c)
		}
	}
}

// login extract the JWT from Authorization header and verify the JWT with firebase credentials.
// It sets the extracted token in the context's field "jwt" and in the "X-Logged-User-ID" request & response headers.
// It first tries to find if the user is already logged (for example, you use the ParseJWT method in a
// general, top level middleware, and the WithRol method in a single endpoint)
func (a *AuthMiddleware) login(c echo.Context) (*auth.Token, error) {
	if userIsAlreadyLogged(c) {
		return GetIDToken(c)
	}
	jwt, err := getJWT(c)
	if err != nil {
		return nil, err
	}
	idToken, err := a.authClient.VerifyIDToken(a.ctx, jwt)
	if err != nil {
		return idToken, err
	}

	setIDToken(c, idToken)
	setUserIDHeader(c, idToken.UID)
	return idToken, err
}

// WithRol searches for a valid JWT with the desired rol as a boolean true key in the token Claims
// and logs in the founded user.
// If no JWT with the rol is found, it returns a 401 or 403 standard error, stopping the request.
func (a *AuthMiddleware) WithRol(rol string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := a.login(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error":   err.Error(),
					"message": "You must log in",
				})
			}

			if !LoggedUserIs(c, rol) {
				return c.JSON(http.StatusForbidden, echo.Map{
					"message": "You have no permission to do this operation",
				})
			}

			return next(c)
		}
	}
}

// WhitAny searches for a valid JWT with at least one of the desired roles as a boolean true
// key in the token Claims and logs in the founded user.
// If no JWT with one desired rol is found, it returns a 401 or 403 standard error, stopping the request.
func (a *AuthMiddleware) WithAny(roles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := a.login(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error":   err.Error(),
					"message": "You must log in",
				})
			}

			if !LoggedUserIsAny(c, roles) {
				return c.JSON(http.StatusForbidden, echo.Map{
					"message": "You have no permission to do this operation",
				})
			}

			return next(c)
		}
	}
}

func getJWT(c echo.Context) (string, error) {
	authorizationHeader := c.Request().Header.Get("Authorization")

	if !strings.HasPrefix(authorizationHeader, bearerPrefix) {
		return "", fmt.Errorf(
			"please, provide an Authorization header whit a %s follow by the JWT token, as %s <jwt>",
			bearerPrefix,
			bearerPrefix,
		)
	}

	jwt := strings.TrimPrefix(authorizationHeader, bearerPrefix)
	return jwt, nil
}
