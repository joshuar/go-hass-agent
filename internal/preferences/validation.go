// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var ErrInternalValidationFailed = errors.New("internal validation error")

func validatePreferences(prefs *Preferences) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(prefs)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

//nolint:err113,errorlint,wsl
func showValidationErrors(e error) error {
	validationErrors, ok := e.(validator.ValidationErrors)
	if !ok {
		return ErrInternalValidationFailed
	}

	var allErrors error

	for _, err := range validationErrors {
		// Namespace:       err.Namespace(),
		// Field:           err.Field(),
		// StructNamespace: err.StructNamespace(),
		// StructField:     err.StructField(),
		// Tag:             err.Tag(),
		// ActualTag:       err.ActualTag(),
		// Kind:            fmt.Sprintf("%v", err.Kind()),
		// Type:            fmt.Sprintf("%v", err.Type()),
		// Value:           fmt.Sprintf("%v", err.Value()),
		// Param:           err.Param(),
		// Message:         err.Error(),

		allErrors = errors.Join(allErrors, fmt.Errorf("%s: %s", err.Field(), err.Error()))
	}

	return allErrors
}
