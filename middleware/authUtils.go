package middleware

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

func initFirebase() {
	if firebaseInitilized {
		return
	}
	var err error
	credential := generateFirebaseCredential()

	ctx := context.Background()
	opt := option.WithCredentialsJSON(credential)
	app, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		panic(fmt.Sprintf("error initializing firebase app: %v\n", err))
	}

	client, err = app.Auth(ctx)
	if err != nil {
		panic(fmt.Sprintf("error initializing firebase client: %v\n", err))
	}
}

type fbCredential struct {
	Type                string
	ProjectID           string
	PrivateKeyID        string
	PrivateKey          string
	ClientEmail         string
	ClientID            string
	AuthURI             string
	TokenURI            string
	AuthProviderCertURL string
	ClientCertURL       string
}

func generateFirebaseCredential() []byte {
	credentialDataTemplate := "{\"type\":\"%s\",\"project_id\":\"%s\",\"private_key_id\":\"%s\",\"private_key\":\"%s\",\"client_email\":\"%s\",\"client_id\":\"%s\",\"auth_uri\":\"%s\",\"token_uri\":\"%s\",\"auth_provider_x509_cert_url\":\"%s\",\"client_x509_cert_url\":\"%s\"}"
	credential := parseFirebaseEnv()
	credentialData := fmt.Sprintf(
		credentialDataTemplate,
		credential.Type,
		credential.ProjectID,
		credential.PrivateKeyID,
		credential.PrivateKey,
		credential.ClientEmail,
		credential.ClientID,
		credential.AuthURI,
		credential.TokenURI,
		credential.AuthProviderCertURL,
		credential.ClientCertURL,
	)

	return []byte(credentialData)
}

const fbTypeEnvKey = "FB_TYPE"
const fbProjectIDEnvKey = "FB_PROJECT_ID"
const fbPrivateKeyIDEnvKey = "FB_PRIVATE_KEY_ID"
const fbPrivateKeyEnvKey = "FB_PRIVATE_KEY"
const fbClientEmailEnvKey = "FB_CLIENT_EMAIL"
const fbClientIDEnvKey = "FB_CLIENT_ID"
const fbAuthURIEnvKey = "FB_AUTH_URI"
const fbTokenURIEnvKey = "FB_TOKEN_URI"
const fbAuthProviderCertURLEnvKey = "FB_AUTH_PROVIDER_CERT_URL"
const fbClientCertURLEnvKey = "FB_CLIENT_CERT_URL"

func parseFirebaseEnv() fbCredential {
	var ok bool
	var credential fbCredential
	credential.Type, ok = os.LookupEnv(fbTypeEnvKey)
	if !ok {
		panicByFbEnv("type", fbTypeEnvKey)
	}

	credential.ProjectID, ok = os.LookupEnv(fbProjectIDEnvKey)
	if !ok {
		panicByFbEnv("project ID", fbProjectIDEnvKey)
	}

	credential.PrivateKeyID, ok = os.LookupEnv(fbPrivateKeyIDEnvKey)
	if !ok {
		panicByFbEnv("private key ID", fbPrivateKeyIDEnvKey)
	}

	credential.PrivateKey, ok = os.LookupEnv(fbPrivateKeyEnvKey)
	if !ok {
		panicByFbEnv("private key", fbPrivateKeyEnvKey)
	}

	credential.ClientEmail, ok = os.LookupEnv(fbClientEmailEnvKey)
	if !ok {
		panicByFbEnv("client email", fbClientEmailEnvKey)
	}

	credential.ClientID, ok = os.LookupEnv(fbClientIDEnvKey)
	if !ok {
		panicByFbEnv("client ID", fbClientIDEnvKey)
	}

	credential.AuthURI, ok = os.LookupEnv(fbAuthURIEnvKey)
	if !ok {
		panicByFbEnv("auth URI", fbAuthURIEnvKey)
	}

	credential.TokenURI, ok = os.LookupEnv(fbTokenURIEnvKey)
	if !ok {
		panicByFbEnv("token URI", fbTokenURIEnvKey)
	}

	credential.AuthProviderCertURL, ok = os.LookupEnv(fbAuthProviderCertURLEnvKey)
	if !ok {
		panicByFbEnv("auth provider cert URL", fbAuthProviderCertURLEnvKey)
	}

	credential.ClientCertURL, ok = os.LookupEnv(fbClientCertURLEnvKey)
	if !ok {
		panicByFbEnv("client cert URL", fbClientCertURLEnvKey)
	}

	return credential
}

func panicByFbEnv(name, key string) {
	panic(fmt.Sprintf("Please specify the firebase credential %s with the environment variable %v", name, key))
}

func setIDToken(c echo.Context, idToken *auth.Token) {
	c.Set(userContextField, idToken)
}

// GetIDToken extract the token from the context field in userContextField and return it.
// If no token is found, it returns a ErrNoIDTokenFound.
func GetIDToken(c echo.Context) (*auth.Token, error) {
	idToken, ok := c.Get(userContextField).(*auth.Token)
	if !ok {
		return nil, ErrNoIDTokenFound
	}
	return idToken, nil
}

var ErrNoIDTokenFound = fmt.Errorf("IDToken not found in context key %s", userContextField)

// LoggedUserIs returns true if a idToken exists in the context and it have the desired rol.
func LoggedUserIs(c echo.Context, rol string) bool {
	idToken, err := GetIDToken(c)
	if err != nil {
		return false
	}

	hasRole, ok := idToken.Claims[rol].(bool)
	return ok && hasRole
}

// LoggedUserIsAny returns true if a idToken exists in the context and it have, at least, one
// of the expecified roles.
func LoggedUserIsAny(c echo.Context, roles []string) bool {
	for _, rol := range roles {
		if LoggedUserIs(c, rol) {
			return true
		}
	}
	return false
}
