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

const bearerPrefix = "bearer "

var app *firebase.App
var client *auth.Client
var firebaseInitilized = false

type AuthMiddleware struct {
	authClient AuthClient
	ctx        context.Context
}

type AuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// DefaultAuthMiddleware return a auth middleware with a firebase auth client initialized from env vars.
func DefaultAuthMiddleware() *AuthMiddleware {
	if !firebaseInitilized {
		initFirebase()
	}
	return NewAuthMiddleware(context.Background(), client)
}

func NewAuthMiddleware(ctx context.Context, authClient AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		ctx:        ctx,
	}
}

// AllowAnonymous will let pass all petitions trying to find a JWT in headers and loging
// in the user if the JWT is found.
func (a *AuthMiddleware) AllowAnonymous() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, _ = a.login(c)
			return next(c)
		}
	}
}

// LoggedUsers searchs for a valid JWT and logs in the founded user. If no JWT is found,
// it returns a 401 unauthorized standar error, stopping the request.
func (a *AuthMiddleware) LoggedUsers() echo.MiddlewareFunc {
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
func (a *AuthMiddleware) login(c echo.Context) (*auth.Token, error) {
	jwt, err := getJWT(c)
	if err != nil {
		return nil, err
	}
	idToken, err := a.authClient.VerifyIDToken(a.ctx, jwt)
	if err == nil {
		setIDToken(c, idToken)
	}
	return idToken, err
}

// WithRol searchs for a valid JWT with the desired rol as a boolean true key in the token Claims
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

// WhitAny searchs for a valid JWT with at least one of the desired roles as a boolean true
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
