// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

//go:generate moq -out mock_Response_test.go . Response
type Response interface {
	SensorRegistrationResponse
	SensorUpdateResponse
	Error() error
	Type() RequestType
}

type SensorRegistrationResponse interface {
	Registered() bool
}

type SensorUpdateResponse interface {
	Disabled() bool
}

type GenericResponse struct {
	error
	requestType RequestType
}

func (e *GenericResponse) Error() error {
	return e.error
}

func (e *GenericResponse) Type() RequestType {
	return e.requestType
}

func (e *GenericResponse) Disabled() bool {
	return false
}

func (e *GenericResponse) Registered() bool {
	return false
}

func NewGenericResponse(e error, t RequestType) *GenericResponse {
	return &GenericResponse{
		error:       e,
		requestType: t,
	}
}
