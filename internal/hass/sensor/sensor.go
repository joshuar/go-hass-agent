// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package sensor

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrHandleSensor = errors.New("error handling sensor data")

type API interface {
	SendRequest(ctx context.Context, url string, req api.Request) (api.Response, error)
	DisableSensor(id models.UniqueID)
}

// newSensorStateRequest takes sensor data and creates a sensor state update
// request.
func newSensorStateRequest(sensor *models.Sensor) (*api.Request, error) {
	req := &api.Request{
		Type:      api.UpdateSensorStates,
		Retryable: sensor.Retryable,
		Data:      &api.Request_Data{},
	}

	state, err := sensor.AsState()
	if err != nil {
		return nil, errors.Join(ErrHandleSensor, err)
	}

	// Add the sensor state into the request.
	err = req.Data.FromSensorState(*state)
	if err != nil {
		return nil, errors.Join(ErrHandleSensor, err)
	}

	return req, nil
}

// newSensorRegistrationRequest takes sensor data and creates a sensor
// registration request.
func newSensorRegistrationRequest(sensor *models.Sensor) (*api.Request, error) {
	req := &api.Request{
		Type:      api.RegisterSensor,
		Retryable: sensor.Retryable,
		Data:      &api.Request_Data{},
	}

	registration, err := sensor.AsRegistration()
	if err != nil {
		return nil, errors.Join(ErrHandleSensor, err)
	}

	// Add the sensor registration into the request.
	err = req.Data.FromSensorRegistration(*registration)
	if err != nil {
		return nil, errors.Join(ErrHandleSensor, err)
	}

	return req, nil
}

// UpdateHandler handles sending sensor data as an update request to Home Assistant and
// processing the response.
func UpdateHandler(ctx context.Context, client API, sensor models.Sensor) error {
	req, err := newSensorStateRequest(&sensor)
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
	}

	resp, err := client.SendRequest(ctx, preferences.RestAPIURL(), *req)
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
	}

	stateResp, err := resp.AsSensorStateResponse()
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
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
				client.DisableSensor(id)
			}
		}
	}

	return nil
}

// RegistrationHandler handles sending sensor data as an registration request to Home Assistant and
// processing the response.
func RegistrationHandler(ctx context.Context, client API, sensor models.Sensor) error {
	req, err := newSensorRegistrationRequest(&sensor)
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
	}

	resp, err := client.SendRequest(ctx, preferences.RestAPIURL(), *req)
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
	}

	registration, err := resp.AsSensorRegistrationResponse()
	if err != nil {
		return errors.Join(ErrHandleSensor, err)
	}

	if registration.Success != nil {
		if !*registration.Success {
			return fmt.Errorf("sensor registration failed")
		}
	}

	return nil
}
