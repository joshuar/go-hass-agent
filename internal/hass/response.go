// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	Registered responseStatus = iota + 1
	Updated
	Disabled
	Failed
)

type responseStatus int

type response struct {
	ErrorDetails *api.ResponseError `json:"error,omitempty"`
	IsSuccess    bool               `json:"success,omitempty"`
}

func (r *response) Status() (responseStatus, error) {
	if r.IsSuccess || r.ErrorDetails == nil {
		return Updated, nil
	}

	return Failed, r.ErrorDetails
}

type sensorUpdateReponse struct {
	response
	IsDisabled bool `json:"is_disabled,omitempty"`
}

func (u *sensorUpdateReponse) Status() (responseStatus, error) {
	switch {
	case !u.IsSuccess:
		return Failed, u.ErrorDetails
	case u.IsDisabled:
		return Disabled, u.ErrorDetails
	default:
		return Updated, nil
	}
}

type bulkSensorUpdateResponse map[string]sensorUpdateReponse

func (u bulkSensorUpdateResponse) Process(ctx context.Context, details sensor.Entity) {
	for id, sensorReponse := range u {
		status, err := sensorReponse.Status()

		switch status {
		case Failed:
			logging.FromContext(ctx).Warn("Sensor update failed.",
				slog.String("id", id),
				slog.Any("error", err))

			return
		case Disabled:
			// Already disabled in registry, nothing to do.
			if sensorRegistry.IsDisabled(id) {
				return
			}
			// Disable in registry.
			logging.FromContext(ctx).
				Info("Sensor is disabled in Home Assistant. Setting disabled in local registry.",
					slog.String("id", id))

			if err := sensorRegistry.SetDisabled(id, true); err != nil {
				logging.FromContext(ctx).Warn("Unable to disable sensor in registry.",
					slog.String("id", id),
					slog.Any("error", err))
			}
		case Updated:
			logging.FromContext(ctx).
				Debug("Sensor updated.",
					sensorLogAttrs(details))
		}

		// Add the sensor update to the tracker.
		if err := sensorTracker.Add(&details); err != nil {
			logging.FromContext(ctx).Warn("Unable to update sensor state in tracker.",
				slog.String("id", id),
				slog.Any("error", err))
		}
	}
}

type sensorRegistrationResponse response

func (r *sensorRegistrationResponse) Status() (responseStatus, error) {
	if r.IsSuccess {
		return Registered, nil
	}

	return Failed, r.ErrorDetails
}

func (r *sensorRegistrationResponse) Process(ctx context.Context, details sensor.Entity) {
	status, err := r.Status()

	switch status {
	case Failed:
		logging.FromContext(ctx).Warn("Sensor registration failed.",
			slog.String("id", details.ID),
			slog.Any("error", err))

		return
	case Registered:
		// Set registration status in registry.
		err = sensorRegistry.SetRegistered(details.ID, true)
		if err != nil {
			logging.FromContext(ctx).Warn("Unable to set sensor registration in registry.",
				slog.String("id", details.ID),
				slog.Any("error", err))
		}
		// Add the sensor update to the tracker.
		if err := sensorTracker.Add(&details); err != nil {
			logging.FromContext(ctx).Warn("Unable to update sensor state in tracker.",
				slog.String("id", details.ID),
				slog.Any("error", err))
		}
	}
}
