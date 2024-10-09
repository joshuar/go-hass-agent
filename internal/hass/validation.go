// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

//nolint:errorlint
//revive:disable:unhandled-error
func parseValidationErrors(validation error) string {
	validationErrs, ok := validation.(validator.ValidationErrors)
	if !ok {
		return "internal validation error"
	}

	var message strings.Builder

	for _, err := range validationErrs {
		switch err.Tag() {
		case "required":
			message.WriteString("field " + err.Field() + " is required")
		default:
			message.WriteString("field " + err.Field() + " should match " + err.Tag())
		}

		message.WriteRune(' ')
	}

	return message.String()
}
