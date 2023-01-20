package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/auth"
	"github.com/labstack/echo/v4"
	"github.com/orov-io/maryread/handler"
	"github.com/stretchr/testify/assert"
)

const mockAuthClientUID = "mockClientUID"
const authTestTrueRol = "Truman"
const authTestFalseRol = "Capote"
const authTestEmptyKeyRol = "TrumanCapote"

func TestDefaultAuthMiddleware(t *testing.T) {
	authMiddleware := DefaultAuthMiddleware()
	assert.NotNil(t, authMiddleware)
}

func TestNewAuthMiddleware(t *testing.T) {
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	assert.NotNil(t, authMiddleware)
}

func TestAllowAnonymousNoJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.AllowAnonymous())

	res, _, _, _ := performAuthPingTestRequest(e, testEmptyJWT)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestAllowAnonymousWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.AllowAnonymous())

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestLoggedUserNoJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.LoggedUser())

	res, _, _, _ := performAuthPingTestRequest(e, testEmptyJWT)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestLoggedUserWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.LoggedUser())

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestWithRoleNoJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithRol(authTestTrueRol))

	res, _, _, _ := performAuthPingTestRequest(e, testEmptyJWT)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestWithRolBadRolWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithRol(authTestFalseRol))

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}

func TestWithRolGoodRolWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithRol(authTestTrueRol))

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestWithAnyNoJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithNoRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithAny([]string{authTestTrueRol, authTestFalseRol}))

	res, _, _, _ := performAuthPingTestRequest(e, testEmptyJWT)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestWithAnyBadRolWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithAny([]string{authTestFalseRol, authTestEmptyKeyRol}))

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}

func TestWithAnyGoodRolWithJWT(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.WithAny([]string{authTestTrueRol, authTestFalseRol}))

	res, _, _, _ := performAuthPingTestRequest(e, testJWT)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUserIDHeaders(t *testing.T) {
	e := echo.New()
	authMiddleware := authMiddlewareWithRolesUserMockClient()
	e.GET(handler.PingPath, handler.NewPingHandler().GetPingHandler, authMiddleware.LoggedUser())

	res, req, _, _ := performAuthPingTestRequest(e, testJWT)
	reqUserID := req.Header.Get(authUserIDHeader)
	resUserID := res.Header.Get(authUserIDHeader)

	assert.Equal(t, mockAuthClientUID, reqUserID)
	assert.Equal(t, mockAuthClientUID, resUserID)
}

// TODO: Test that Parse as general middleware works fine with another middleware at request level.

func performAuthPingTestRequest(e *echo.Echo, jwt string) (res *http.Response, req *http.Request, data []byte, err error) {
	req = httptest.NewRequest(http.MethodGet, handler.PingPath, nil)
	rec := httptest.NewRecorder()
	if jwt != "" {
		req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("%v %v", testJWTHeaderPrefix, jwt))
	}

	e.ServeHTTP(rec, req)
	res = rec.Result()
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	return
}

func authMiddlewareWithNoRolesUserMockClient() *AuthMiddleware {
	mockClient := newMockAuthClient()
	mockClient.Token.UID = mockAuthClientUID
	return NewAuthMiddleware(context.Background(), mockClient)
}

func authMiddlewareWithRolesUserMockClient() *AuthMiddleware {
	mockClient := newMockAuthClient()
	mockClient.Token.UID = mockAuthClientUID
	mockClient.Token.Claims = map[string]interface{}{authTestTrueRol: true, authTestFalseRol: false}
	return NewAuthMiddleware(context.Background(), mockClient)
}

type mockAuthClient struct {
	Token *auth.Token
}

func newMockAuthClient() *mockAuthClient {
	return &mockAuthClient{
		Token: new(auth.Token),
	}
}

func (m *mockAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return m.Token, nil
}
