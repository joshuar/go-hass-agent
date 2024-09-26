// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import "fmt"

type apiError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("code %v: %s", e.Code, e.Message)
}

type responseStatus struct {
	ErrorDetails *apiError
	IsSuccess    bool `json:"success,omitempty"`
}

type updateResponseStatus struct {
	responseStatus
	IsDisabled bool `json:"is_disabled,omitempty"`
}

func (u *updateResponseStatus) disabled() bool {
	return u.IsDisabled
}

func (u *updateResponseStatus) success() (bool, error) {
	if u.IsSuccess {
		return true, nil
	}

	return false, u.ErrorDetails
}

type stateUpdateResponse map[string]updateResponseStatus

type registrationResponse responseStatus

func (r *registrationResponse) registered() (bool, error) {
	if r.IsSuccess {
		return true, nil
	}

	return false, r.ErrorDetails
}

type locationResponse struct {
	error
}

//nolint:staticcheck
func (r *locationResponse) updated() error {
	return r
}
