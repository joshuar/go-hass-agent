// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func validatePreferences(prefs *Preferences) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(prefs)
}

func showValidationErrors(e error) error {
	validationErrors, ok := e.(validator.ValidationErrors)
	if !ok {
		return errors.New("unable to assert as validation errors")
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

		log.Error().
			Str("preference", err.Field()).
			Err(errors.New(err.Error())).
			Msg("Validation failed.")
		allErrors = errors.Join(allErrors, fmt.Errorf("%s: %s", err.Field(), err.Error()))
	}
	return allErrors
}
