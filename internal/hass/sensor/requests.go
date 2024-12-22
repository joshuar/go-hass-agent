// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package sensor

const (
	requestTypeRegisterSensor = "register_sensor"
	requestTypeUpdateSensor   = "update_sensor_states"
	requestTypeLocation       = "update_location"
)

// Request represents a sensor request, either a registration, update or
// location update.
type Request struct {
	Data        any    `json:"data"`
	RequestType string `json:"type"`
	retryable   bool
}

func (r *Request) RequestBody() any {
	return r
}

func (r *Request) Retry() bool {
	return r.retryable
}

// AsSensorUpdate indicates the request will be a sensor update request.
func AsSensorUpdate(entity Entity) Option[Request] {
	return func(request Request) Request {
		request.RequestType = requestTypeUpdateSensor
		request.Data = &struct {
			State      any            `json:"state" validate:"required"`
			Attributes map[string]any `json:"attributes,omitempty" validate:"omitempty"`
			Icon       string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
			ID         string         `json:"unique_id" validate:"required"`
			EntityType string         `json:"type" validate:"omitempty"`
		}{
			State:      entity.Value,
			Attributes: entity.Attributes,
			Icon:       entity.Icon,
			ID:         entity.ID,
			EntityType: entity.EntityType.String(),
		}

		return request
	}
}

// AsSensorRegistration indicates the request will be a sensor registration
// request.
func AsSensorRegistration(entity Entity) Option[Request] {
	return func(request Request) Request {
		request.RequestType = requestTypeRegisterSensor
		request.Data = &struct {
			State       any            `json:"state" validate:"required"`
			Attributes  map[string]any `json:"attributes,omitempty" validate:"omitempty"`
			Icon        string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
			ID          string         `json:"unique_id" validate:"required"`
			EntityType  string         `json:"type" validate:"omitempty"`
			Name        string         `json:"name" validate:"required"`
			Units       string         `json:"unit_of_measurement,omitempty" validate:"omitempty"`
			DeviceClass string         `json:"device_class,omitempty" validate:"omitempty"`
			StateClass  string         `json:"state_class,omitempty" validate:"omitempty"`
			Category    string         `json:"entity_category,omitempty" validate:"omitempty"`
		}{
			State:       entity.Value,
			Attributes:  entity.Attributes,
			Icon:        entity.Icon,
			ID:          entity.ID,
			EntityType:  entity.EntityType.String(),
			Name:        entity.Name,
			Units:       entity.Units,
			DeviceClass: entity.DeviceClass.String(),
			StateClass:  entity.StateClass.String(),
			Category:    entity.Category.String(),
		}

		return request
	}
}

// AsLocationUpdate indicates the request will be a location update request.
func AsLocationUpdate(location Location) Option[Request] {
	return func(request Request) Request {
		request.RequestType = requestTypeLocation
		request.Data = &location

		return request
	}
}

// AsRetryable marks that the request should be retried.
func AsRetryable(value bool) Option[Request] {
	return func(request Request) Request {
		request.retryable = value
		return request
	}
}

// NewRequest creates a new request with the given options.
func NewRequest(options ...Option[Request]) *Request {
	request := Request{}

	for _, option := range options {
		request = option(request)
	}

	return &request
}
