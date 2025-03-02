// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ErrValidationFailed = errors.New("internal validation error")

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidationErrors is a map of fields and their validation errors.
type ValidationErrors map[string][]string

func (p ValidationErrors) Error() string {
	var message strings.Builder
	for field, problems := range p {
		// Write the field name.
		message.WriteString(fmt.Sprintf("field %s: ", field))
		// Write each problem with the field.
		for _, problem := range problems {
			message.WriteString(problem)
		}

		message.WriteRune(' ')
	}

	return message.String()
}

//nolint:errorlint
//revive:disable:unhandled-error
func ParseValidationErrors(validation error) string {
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

func parseStructValidationErrors(validationErrors validator.ValidationErrors) ValidationErrors {
	problems := make(ValidationErrors)

	for _, err := range validationErrors {
		field := err.Field()

		problems[field] = append(problems[field], fmt.Sprintf("%s: %s", err.Tag(), err.Error()))
	}

	return problems
}

//nolint:errorlint,errcheck
func ValidateStruct[T any](obj T) (bool, ValidationErrors) {
	validationErr := &validator.ValidationErrors{}

	err := Validate.Struct(obj)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, map[string][]string{
				"internal": {ErrValidationFailed.Error()},
			}
		}

		problems := parseStructValidationErrors(err.(validator.ValidationErrors))

		return false, problems
	}

	return true, nil
}

func ValidateVariable(variable any, rule string) (bool, error) {
	if err := Validate.Var(variable, rule); err != nil {
		return false, fmt.Errorf("invalid: %w", err)
	}

	return true, nil
}
