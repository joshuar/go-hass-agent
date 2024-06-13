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

//nolint:err113,errorlint
func showValidationErrors(e error) error {
	validationErrors, ok := e.(validator.ValidationErrors)
	if !ok {
		return ErrInternalValidationFailed
	}

	var allErrors error

	for _, err := range validationErrors {
		allErrors = errors.Join(allErrors, fmt.Errorf("%s: %s", err.Field(), err.Error()))
	}

	return allErrors
}
