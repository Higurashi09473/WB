// Package validator provides validation utilities for domain models.
// It uses go-playground/validator under the hood.
package validator

import (
	"WB/internal/models"

	"github.com/go-playground/validator/v10"
)

// validate is a global instance of the validator.
// Initialized once during application startup.
var validate *validator.Validate

// init initializes the global validator instance.
// Called automatically when the package is imported.
func init() {
    validate = validator.New()
}

// ValidateOrder validates the Order model using predefined struct tags.
// Returns nil if the order is valid, or an error describing validation failures.
func ValidateOrder(order *models.Order) error {
    return validate.Struct(order)
}