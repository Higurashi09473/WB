package validator

import (
	"WB/internal/models"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
}

func ValidateOrder(order *models.Order) error {
    return validate.Struct(order)
}