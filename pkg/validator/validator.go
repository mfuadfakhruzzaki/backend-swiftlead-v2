package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator is a wrapper around the validator library
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// Global validator instance
var Default = New()

// Validate is a helper to validate a struct using the default validator
func Validate(s interface{}) error {
	return Default.ValidateStruct(s)
}
