package maryread

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

// Validator is the default validator. It uses the go-playground validator and wraps
// the error in a echo.HTTPError with a status 400 error.
// For more control over the status code, please use the ValidatorRawError
type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// ValidatorRawError uses the go-playground validator and return the raw error directly. No handling this
// error will become in a InternalServerError (500) if validation fails and you return the error to echo.
type ValidatorRawError struct {
	validator *validator.Validate
}

func NewValidatorRawError() *ValidatorRawError {
	return &ValidatorRawError{
		validator: validator.New(),
	}
}

func (v *ValidatorRawError) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// Bindlidate uses provided context to both Bind and Validate data.
func Bindlidate[T any](c echo.Context, data *T) error {
	err := c.Bind(data)
	if err != nil {
		return err
	}

	err = c.Validate(data)
	return err
}
