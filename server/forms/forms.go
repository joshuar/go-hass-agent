// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package forms contains methods for handling form decoding and encoding.
package forms

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-playground/form/v4"
)

var (
	// ErrDecode indicates an error occurred during decoding.
	ErrDecode = errors.New("error in decoding")
	// ErrEncode indicates an error occurred during encoding.
	ErrEncode = errors.New("error in encoding")
	// ErrValidation indicates an error occurred during validation.
	ErrValidation = errors.New("validation failed")
	// ErrSanitise indicates an error occurred during sanitisation.
	ErrSanitise = errors.New("sanitisation failed")
)

var (
	decoder = form.NewDecoder()
	encoder = form.NewEncoder()
)

// FormInput represents form input data. It has methods to test if the data is valid and to sanitise the input data.
type FormInput interface {
	Valid() (bool, error)
	Sanitise() error
}

// DecodeForm will decode submitted form contents into the passed in type. It
// will perform validation of the type and will return the type and a boolean
// true if it is valid. If decoding the form submission fails, a non-nill error
// is returned.
func DecodeForm[T FormInput](req *http.Request) (T, bool, error) {
	var obj T
	// Parse form values in request.
	if err := req.ParseForm(); err != nil {
		return obj, false, fmt.Errorf("%w: %w", ErrDecode, err)
	}
	// Decode the form values.
	err := decoder.Decode(&obj, req.Form)
	if err != nil {
		return obj, false, fmt.Errorf("%w: %w", ErrDecode, err)
	}
	// Sanitise the object.
	if err := obj.Sanitise(); err != nil {
		return obj, false, fmt.Errorf("%w: %w", ErrSanitise, err)
	}
	// Validate the object.
	if ok, err := obj.Valid(); !ok {
		return obj, false, fmt.Errorf("%w: %w", ErrValidation, err)
	}
	return obj, true, nil
}

// EncodeForm will encode the given object as url.Values, using the struct tags
// where possible. It will perform validation of the object before attempting
// encoding. If the object cannot be encoded or validation fails, a non-nil
// error is returned.
func EncodeForm[T FormInput](obj T) (url.Values, error) {
	// Sanitise the object.
	if err := obj.Sanitise(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSanitise, err)
	}
	// Validate the object.
	if ok, err := obj.Valid(); !ok {
		return nil, fmt.Errorf("%w: %w", ErrValidation, err)
	}
	values, err := encoder.Encode(&obj)
	if err != nil {
		return nil, errors.Join(ErrEncode, err)
	}
	return values, nil
}
