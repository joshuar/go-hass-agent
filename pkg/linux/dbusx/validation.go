// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package dbusx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrValidationFailed indicates that the validation action itself failed.
	ErrValidationFailed = errors.New("internal validation error")
	// ErrValidation indicates that validation did not pass.
	ErrValidation = errors.New("validation failed")

	ErrInvalidField  = errors.New("invalid field")
	ErrInvalidStruct = errors.New("invalid struct")
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// FieldError is a particular validation error on a particular field.
type FieldError struct {
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

// Error satisfies the Error interface.
func (e *FieldError) Error() string {
	return fmt.Sprintf(
		"%s %s (value(%q)) failed validation for %s: %s",
		ErrInvalidField.Error(),
		e.Field,
		e.Value,
		e.Tag,
		e.Message,
	)
}

// StructError contains validation errors on individual fields in a struct.
type StructError struct {
	Fields []FieldError
}

// Error satisfies the Error interface.
func (e *StructError) Error() string {
	var errStr strings.Builder
	errStr.WriteString("contains field errors")
	if len(e.Fields) > 0 {
		errStr.WriteRune('\n')
	}
	for idx, t := range e.Fields {
		errStr.WriteString(t.Error())
		if idx < (len(e.Fields) - 1) {
			errStr.WriteRune('\n')
		}
	}
	return errStr.String()
}

func valid[T any](obj *T) error {
	if err := validate.Struct(obj); err != nil {
		errs := &StructError{}
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			errs.Fields = make([]FieldError, 0, len(validateErrs))
			for _, err := range validateErrs {
				errs.Fields = append(errs.Fields, FieldError{
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
				})
			}
			return errs
		}
	}
	return nil
}
