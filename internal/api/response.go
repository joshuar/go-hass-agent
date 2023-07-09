// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

//go:generate mockery --name Response --inpackage
type Response interface {
	SensorRegistrationResponse
	SensorUpdateResponse
	Error() error
	Type() RequestType
}

//go:generate mockery --name SensorRegistrationResponse --inpackage
type SensorRegistrationResponse interface {
	Registered() bool
}

//go:generate mockery --name SensorUpdateResponse --inpackage
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
