package api

import (
	"github.com/go-playground/validator/v10"                    // Import the validator library for custom validations
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util" // Import utility functions
)

// ---------------------------
// Custom Currency Validator
// ---------------------------

// validCurrency is a custom validator function that checks whether a string
// represents a supported currency in the system. It will be used in request
// binding tags like `binding:"required,currency"` or Fiber's equivalent.
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// Attempt to get the field's value as a string
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// Use the utility function IsSupportedCurrency to check validity
		return util.IsSupportedCurrency(currency)
	}

	// If the field is not a string, validation fails
	return false
}
