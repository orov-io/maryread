package maryread

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type testValidatorUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

const (
	testValidatorValidName    = "Truman"
	testValidatorEmptyName    = ""
	testValidatorValidEmail   = "truman@capote.com"
	testValidatorEmptyEmail   = ""
	testValidatorInvalidEmail = "truman"

	testValidatorhandlerPath       = "/validator"
	testValidatorRequestValidURI   = "/validator?id=1"
	testValidatorRequestInvalidURI = "/validator?id=truman"
)

var (
	userFailAll = testValidatorUser{
		Email: testValidatorInvalidEmail,
	}

	userFailInName = testValidatorUser{
		Name:  testValidatorEmptyName,
		Email: testValidatorValidEmail,
	}

	emailFailEmpty = testValidatorUser{
		Name:  testValidatorValidName,
		Email: testValidatorEmptyEmail,
	}

	emailFailInvalid = testValidatorUser{
		Name:  testValidatorEmptyName,
		Email: testValidatorInvalidEmail,
	}

	testValidatorValidUser = testValidatorUser{
		Name:  testValidatorValidName,
		Email: testValidatorValidEmail,
	}
)

func TestCustomValidator(t *testing.T) {
	validator := NewValidator()

	err := validator.Validate(userFailAll)
	assert.Error(t, err)
	assert.IsType(t, &echo.HTTPError{}, err)

	err = validator.Validate(userFailInName)
	assert.Error(t, err)
	assert.IsType(t, &echo.HTTPError{}, err)

	err = validator.Validate(emailFailEmpty)
	assert.Error(t, err)
	assert.IsType(t, &echo.HTTPError{}, err)

	err = validator.Validate(emailFailInvalid)
	assert.Error(t, err)
	assert.IsType(t, &echo.HTTPError{}, err)

	err = validator.Validate(testValidatorValidUser)
	assert.NoError(t, err)
}

func TestCustomValidatorRawError(t *testing.T) {
	validator := NewValidatorRawError()

	err := validator.Validate(userFailAll)
	assert.Error(t, err)

	err = validator.Validate(userFailInName)
	assert.Error(t, err)

	err = validator.Validate(emailFailEmpty)
	assert.Error(t, err)

	err = validator.Validate(emailFailInvalid)
	assert.Error(t, err)

	err = validator.Validate(testValidatorValidUser)
	assert.NoError(t, err)
}

func TestBindlidate_ValidQuery(t *testing.T) {
	e := echo.New()
	e.Validator = NewValidator()

	req := httptest.NewRequest(http.MethodGet, testValidatorRequestValidURI, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	query := new(testValidatorQueryRequest)
	err := Bindlidate(c, query)

	assert.NoError(t, err)
}

func TestBindlidate_InvalidQuery(t *testing.T) {
	e := echo.New()
	e.Validator = NewValidator()

	req := httptest.NewRequest(http.MethodGet, testValidatorRequestInvalidURI, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	query := new(testValidatorQueryRequest)
	err := Bindlidate(c, query)

	assert.Error(t, err)
}

type testValidatorQueryRequest struct {
	Id int `query:"id" validate:"required,number"`
}
