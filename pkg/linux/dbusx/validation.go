// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package dbusx

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type validationError struct {
	Namespace       string `json:"namespace"` // can differ when a custom TagNameFunc is registered or
	Field           string `json:"field"`     // by passing alt name to ReportError like below
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
}

var ErrValidationError = errors.New("internal validation error")

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func getValidationProblems(validationErrors validator.ValidationErrors) map[string]string {
	problems := make(map[string]string)

	for _, err := range validationErrors {
		errInfo := validationError{
			Namespace:       err.Namespace(),
			Field:           err.Field(),
			StructNamespace: err.StructNamespace(),
			StructField:     err.StructField(),
			Tag:             err.Tag(),
			ActualTag:       err.ActualTag(),
			Kind:            fmt.Sprintf("%v", err.Kind()),
			Type:            fmt.Sprintf("%v", err.Type()),
			Value:           fmt.Sprintf("%v", err.Value()),
			Param:           err.Param(),
			Message:         err.Error(),
		}

		problems[errInfo.Field] = errInfo.Message
	}

	// from here you can create your own error messages in whatever language you wish
	return problems
}

//nolint:errorlint,forcetypeassert
func valid[T any](obj *T) error {
	err := validate.Struct(obj)
	if err != nil {
		switch {
		case errors.Is(err, &validator.InvalidValidationError{}):
			return ErrValidationError
		case errors.Is(err, validator.ValidationErrors{}):
			var errs error
			for field, problem := range getValidationProblems(err.(validator.ValidationErrors)) {
				errs = errors.Join(errs, fmt.Errorf("%s: %s", field, problem)) //nolint:err113
			}

			return errs
		}
	}

	return nil
}
