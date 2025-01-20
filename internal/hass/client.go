// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	DefaultTimeout = 30 * time.Second
)

var (
	ErrGetConfigFailed   = errors.New("could not fetch Home Assistant config")
	ErrGenRequestFailed  = errors.New("unable to generate request for sensor")
	ErrSendRequestFailed = errors.New("could not send sensor request to Home Assistant")

	ErrStateUpdateUnknown       = errors.New("unknown sensor update response")
	ErrStateUpdateFailed        = errors.New("state update failed")
	ErrRegDisableFailed         = errors.New("failed to disable sensor in registry")
	ErrRegAddFailed             = errors.New("failed to set registered status for sensor in registry")
	ErrTrkUpdateFailed          = errors.New("failed to update sensor state in tracker")
	ErrSensorRegistrationFailed = errors.New("sensor registration failed")

	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidClient     = errors.New("invalid client")
	ErrResponseMalformed = errors.New("malformed response")
	ErrUnknown           = errors.New("unknown error occurred")

	ErrInvalidSensor = errors.New("invalid sensor")
)

// sensorRegistry represents the required methods for hass to manage sensor
// registration state.
type sensorRegistry interface {
	SetDisabled(id string, state bool) error
	SetRegistered(id string, state bool) error
	IsDisabled(id string) bool
	IsRegistered(id string) bool
}

// sensorTracker represents the required methods for hass to track sensors and
// their current state.
type sensorTracker interface {
	SensorList() []string
	Get(id string) (*sensor.Entity, error)
	Add(details *sensor.Entity) error
}

type handler struct {
	logger   *slog.Logger
	registry sensorRegistry
	tracker  sensorTracker
}

func NewDataHandler(ctx context.Context, reg sensorRegistry, trk sensorTracker) (chan any, error) {
	dataCh := make(chan any)

	client := &handler{
		logger:   logging.FromContext(ctx).With(slog.String("subsystem", "hass")),
		registry: reg,
		tracker:  trk,
	}

	go func() {
		for d := range dataCh {
			var err error
			switch data := d.(type) {
			case sensor.Entity:
				err = client.processSensor(ctx, data)
			case event.Event:
				err = client.processEvent(ctx, data)
			}

			if err != nil {
				client.logger.Error("Processing failed.", slog.Any("error", err))
			}
		}
	}()

	return dataCh, nil
}

func (c *handler) processEvent(ctx context.Context, details event.Event) error {
	resp, err := api.Send[response](ctx, preferences.RestAPIURL(), &details)
	if err != nil {
		return fmt.Errorf("failed to send event request: %w", err)
	}

	if _, err := resp.Status(); err != nil {
		return err
	}

	c.logger.Debug("Event sent.",
		eventLogAttrs(details))

	return nil
}

func (c *handler) processSensor(ctx context.Context, details sensor.Entity) error {
	// Location request.
	if req, ok := details.Value.(*sensor.Location); ok {
		resp, err := api.Send[response](ctx, preferences.RestAPIURL(),
			sensor.NewRequest(
				sensor.AsLocationUpdate(*req),
			))
		if err != nil {
			return fmt.Errorf("failed to send location update: %w", err)
		}

		if _, err := resp.Status(); err != nil {
			return err
		}

		return nil
	}

	// Sensor update.
	if c.registry.IsRegistered(details.ID) {
		// For sensor updates, if the sensor is disabled, don't continue.
		if c.isDisabled(ctx, details) {
			c.logger.
				Debug("Not sending request for disabled sensor.",
					sensorLogAttrs(details))

			return nil
		}

		resp, err := api.Send[bulkSensorUpdateResponse](ctx, preferences.RestAPIURL(),
			sensor.NewRequest(
				sensor.AsSensorUpdate(details),
				sensor.AsRetryable(details.RetryRequest),
			))
		if err != nil {
			return fmt.Errorf("failed to send sensor update for %s: %w", details.Name, err)
		}

		go resp.Process(ctx, c.registry, c.tracker, details)

		return nil
	}

	// Sensor registration.
	resp, err := api.Send[sensorRegistrationResponse](ctx, preferences.RestAPIURL(),
		sensor.NewRequest(
			sensor.AsSensorRegistration(details),
			sensor.AsRetryable(details.RetryRequest),
		))
	if err != nil {
		return fmt.Errorf("failed to send sensor registration: %w", err)
	}

	go resp.Process(ctx, c.registry, c.tracker, details)

	return nil
}

// isDisabled handles processing a sensor that is disabled. For a sensor that is
// disabled, we need to make an additional check against Home Assistant to see
// if the sensor has been re-enabled, and update our local registry before
// continuing.
func (c *handler) isDisabled(ctx context.Context, details sensor.Entity) bool {
	// If it is not disabled in the local registry, immediately return false.
	if !c.isDisabledInReg(details.ID) {
		return false
	}
	// Else, get the disabled state from Home Assistant
	disabledInHA := c.isDisabledInHA(ctx, details)

	// If sensor is no longer disabled in Home Assistant, update the local
	// registry and return false.
	if !disabledInHA {
		c.logger.Info("Sensor re-enabled in Home Assistant, Re-enabling in local registry and sending updates.",
			sensorLogAttrs(details))

		if err := c.registry.SetDisabled(details.ID, false); err != nil {
			c.logger.Error("Could not re-enable sensor.",
				sensorLogAttrs(details),
				slog.Any("error", err))

			return true
		}

		return false
	}

	// Sensor is disabled in both the local registry and Home Assistant.
	// Return true.
	return true
}

// isDisabledInReg returns the disabled state of the sensor from the local
// registry.
//
//revive:disable:unused-receiver
func (c *handler) isDisabledInReg(id string) bool {
	return c.registry.IsDisabled(id)
}

// isDisabledInHA returns the disabled state of the sensor from Home Assistant.
func (c *handler) isDisabledInHA(ctx context.Context, details sensor.Entity) bool {
	config, err := api.Send[Config](ctx, preferences.RestAPIURL(), &configRequest{})
	if err != nil {
		c.logger.
			Debug("Could not fetch Home Assistant config. Assuming sensor is still disabled.",
				sensorLogAttrs(details),
				slog.Any("error", err))

		return true
	}

	status, err := config.IsEntityDisabled(details.ID)
	if err != nil {
		c.logger.
			Debug("Could not determine sensor disabled status in Home Assistant config. Assuming sensor is still disabled.",
				sensorLogAttrs(details),
				slog.Any("error", err))

		return true
	}

	return status
}

// sensorLogAttrs is a convienience function that returns some slog attributes
// for priting sensor details in the log.
func sensorLogAttrs(details sensor.Entity) slog.Attr {
	return slog.Group("sensor",
		slog.String("name", details.Name),
		slog.String("id", details.ID),
		slog.Any("state", details.Value),
		slog.String("units", details.Units),
	)
}

func eventLogAttrs(details event.Event) slog.Attr {
	return slog.Group("event",
		slog.String("type", details.EventType),
	)
}
