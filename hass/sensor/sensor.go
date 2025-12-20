// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package sensor

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/validation"
)

var ErrHandleSensor = errors.New("error handling sensor data")

type API interface {
	SendRequest(ctx context.Context, url string, req api.RequestData) (api.ResponseData, error)
	DisableSensor(ctx context.Context, id models.UniqueID)
	RestAPIURL() string
}

// newSensorRequest takes sensor data and creates an api.Request for the given
// request type.
func newSensorRequest(reqType api.RequestType, sensor *models.Sensor) (*api.RequestData, error) {
	if valid, problems := validation.ValidateStruct(sensor); !valid {
		return nil, fmt.Errorf("%w: %w", ErrHandleSensor, problems)
	}

	req := &api.RequestData{
		Type:      reqType,
		Retryable: sensor.Retryable,
		Payload:   api.RequestData_Payload{},
	}

	switch reqType {
	case api.UpdateSensorStates:
		state, err := sensor.AsState()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrHandleSensor, err)
		}

		// Add the sensor state into the request.
		err = req.Payload.FromSensorState(*state)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrHandleSensor, err)
		}
	case api.RegisterSensor:
		registration, err := sensor.AsRegistration()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrHandleSensor, err)
		}

		// Add the sensor registration into the request.
		err = req.Payload.FromSensorRegistration(*registration)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrHandleSensor, err)
		}
	}

	return req, nil
}

// UpdateHandler handles sending sensor data as an update request to Home Assistant and
// processing the response.
func UpdateHandler(ctx context.Context, client API, sensor models.Sensor) error {
	req, err := newSensorRequest(api.UpdateSensorStates, &sensor)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	resp, err := client.SendRequest(ctx, client.RestAPIURL(), *req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	stateResp, err := resp.AsSensorStateResponse()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	for id, status := range stateResp {
		if err := status.HasError(); err != nil {
			return fmt.Errorf("sensor update failed for %s: %w", id, err)
		}

		success, err := status.HasSuccess()
		if err != nil {
			return fmt.Errorf("indeterminate status response for sensor %s: %w", id, err)
		}

		if !success {
			return fmt.Errorf("sensor update was unsuccessful %s: %w", id, err)
		}

		if success {
			if status.SensorDisabled() {
				// If the response indicates the sensor has been disabled in
				// Home Assistant, also disable in the local registry.
				client.DisableSensor(ctx, id)
			}
		}
	}

	return nil
}

// RegistrationHandler handles sending sensor data as an registration request to Home Assistant and
// processing the response.
func RegistrationHandler(ctx context.Context, client API, sensor models.Sensor) (bool, error) {
	req, err := newSensorRequest(api.RegisterSensor, &sensor)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	resp, err := client.SendRequest(ctx, client.RestAPIURL(), *req)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	registration, err := resp.AsSensorRegistrationResponse()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrHandleSensor, err)
	}

	if !registration.Success {
		return false, fmt.Errorf("%w: sensor registration failed", ErrHandleSensor)
	}

	return true, nil
}
