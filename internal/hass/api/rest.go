// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package api

import (
	"errors"
	"strings"
)

var (
	ErrUnknownResponse     = errors.New("an unknown error was returned in the response")
	ErrUnspecifiedResponse = errors.New("an unspecified error was returned in the response")
)

// Error allows an APIError to satisfy the Go Error interface.
func (e *APIError) Error() string {
	var msg []string
	if e.Code != nil {
		msg = append(msg, *e.Code)
	}

	if e.Message != nil {
		msg = append(msg, *e.Message)
	}

	if len(msg) == 0 {
		msg = append(msg, "unknown error")
	}

	return strings.Join(msg, ": ")
}

// HasError determines whether the response status indicates an error condition
// has occurred.
func (r *ResponseStatus) HasError() error {
	if r.Error.IsSpecified() {
		apiErr, err := r.Error.Get()
		if err != nil {
			return errors.Join(ErrUnknownResponse, err)
		}

		return &apiErr
	}

	if r.Error.IsNull() {
		return ErrUnspecifiedResponse
	}

	return nil
}

// HasSuccess determines whether the response status indicates the request was
// successful.
func (r *ResponseStatus) HasSuccess() (bool, error) {
	if r.IsSuccess.IsSpecified() {
		return r.IsSuccess.Get()
	}

	if r.IsSuccess.IsNull() {
		return true, nil
	}

	return false, nil
}

// SensorDisabled determines whether the response status indicates the sensor was disabled.
func (r *ResponseStatus) SensorDisabled() bool {
	if r.IsDisabled.IsSpecified() {
		if disabled, err := r.IsDisabled.Get(); err == nil {
			return disabled
		}
	}

	return false
}
