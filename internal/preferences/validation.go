// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

var ErrInternalValidationFailed = errors.New("internal validation error")

//nolint:tagliatelle
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

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

//nolint:errorlint
func showValidationErrors(validation error) {
	validationErrs, ok := validation.(validator.ValidationErrors)
	if !ok {
		slog.Error("Validation error.", "error", ErrInternalValidationFailed)

		return
	}

	for _, err := range validationErrs {
		errDetails := validationError{
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

		indent, err := json.MarshalIndent(errDetails, "", "  ")
		if err != nil {
			slog.Error("Validation error.", "error", err.Error())
			panic(err)
		}

		slog.Error("Validation", "error", string(indent))
	}
}
