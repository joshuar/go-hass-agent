// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/hass/location"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/tracker"
	"github.com/joshuar/go-hass-agent/internal/models"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultRetryWait    = 5 * time.Second
	defaultRetryCount   = 5
	defaultRetryMaxWait = 20 * time.Second
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
	SensorList() []models.UniqueID
	Get(id models.UniqueID) (*models.Sensor, error)
	Add(details *models.Sensor) error
}

// Client handles incoming entity data from the agent and sends appropriate
// requests to the Home Assistant API(s).
type Client struct {
	sensorRegistry sensorRegistry
	sensorTracker  sensorTracker
	restAPI        *resty.Client
	logger         *slog.Logger
}

var (
	ErrNewRequest  = errors.New("could not create request")
	ErrSendRequest = errors.New("send request failed")
	ErrGetConfig   = errors.New("error retrieving Home Assistant config")
)

// isDisabled handles processing a sensor that is disabled. For a sensor that is
// disabled, we need to make an additional check against Home Assistant to see
// if the sensor has been re-enabled, and update our local registry before
// continuing.
func (c *Client) isDisabled(ctx context.Context, details models.Sensor) bool {
	regDisabled := c.isDisabledInReg(details.UniqueID)
	haDisabled := c.isDisabledInHA(ctx, details)

	switch {
	case regDisabled && !haDisabled:
		c.logger.Info("Sensor re-enabled in Home Assistant, Re-enabling in local registry and sending updates.",
			details.LogAttributes())
		c.sensorRegistry.SetDisabled(details.UniqueID, false)
		return false
	case !regDisabled && haDisabled:
		c.logger.Info("Sensor has been disabled in Home Assistant, Disabling in local registry and not sending updates.",
			details.LogAttributes())
		c.sensorRegistry.SetDisabled(details.UniqueID, true)
		return true
	case regDisabled && haDisabled:
		c.logger.Info("Sensor is disabled, not sending updates.",
			details.LogAttributes())
		return true
	}

	return false
}

// isDisabledInReg returns the disabled state of the sensor from the local
// registry.
//
//revive:disable:unused-receiver
func (c *Client) isDisabledInReg(id string) bool {
	return c.sensorRegistry.IsDisabled(id)
}

// isDisabledInHA returns the disabled state of the sensor from Home Assistant.
func (c *Client) isDisabledInHA(ctx context.Context, sensor models.Sensor) bool {
	config, err := c.getHAConfig(ctx)
	if err != nil {
		c.logger.Debug("Could not retrieve Home Assistant config. Assuming sensor is NOT disabled.",
			sensor.LogAttributes(),
			slog.Any("error", err))
		return false
	}

	status, err := config.IsEntityDisabled(sensor.UniqueID)
	if err != nil {
		c.logger.Debug("Could not retrieve Home Assistant config. Assuming sensor is NOT disabled.",
			slog.Any("error", err))
		return false
	}

	return status
}

// GetHAConfig retrieves the Home Assistant config.
func (c *Client) getHAConfig(ctx context.Context) (*api.ConfigResponse, error) {
	resp, err := c.SendRequest(ctx, preferences.RestAPIURL(), api.Request{Type: api.GetConfig})
	if err != nil {
		return nil, errors.Join(ErrGetConfig, err)
	}

	config, err := resp.AsConfigResponse()
	if err != nil {
		return nil, errors.Join(ErrGetConfig, err)
	}

	return &config, nil
}

// EntityHandler takes incoming Entity objects via the passed in channel and
// runs the appropriate handler for the Entity type.
func (c *Client) EntityHandler(ctx context.Context, entityCh chan models.Entity) {
	ctx = logging.ToContext(ctx, c.logger)

	for entity := range entityCh {
		if eventData, err := entity.AsEvent(); err == nil && eventData.Valid() {
			// Send event.
			if err := event.Handler(ctx, c, eventData); err != nil {
				c.logger.Warn("Could not send event.",
					eventData.LogAttributes(),
					slog.Any("error", err))
			}

			continue
		}

		if locationData, err := entity.AsLocation(); err == nil && locationData.Valid() {
			// Send location update.
			if err := location.Handler(ctx, c, locationData); err != nil {
				c.logger.Warn("Could not update location.",
					slog.Any("error", err))
			}

			continue
		}

		if sensorData, err := entity.AsSensor(); err == nil {
			// Send sensor details.
			if c.sensorRegistry.IsRegistered(sensorData.UniqueID) && !c.isDisabled(ctx, sensorData) {
				// If the sensor is registered and not disabled, send an update request.
				if err := sensor.UpdateHandler(ctx, c, sensorData); err != nil {
					c.logger.Warn("Could not update sensor.",
						sensorData.LogAttributes(),
						slog.Any("error", err))

					continue
				}
				c.logger.Debug("Sensor updated.",
					sensorData.LogAttributes())
			} else {
				// Otherwise, send a registration request.
				if err := sensor.RegistrationHandler(ctx, c, sensorData); err != nil {
					c.logger.Warn("Could not register sensor.",
						sensorData.LogAttributes(),
						slog.Any("error", err))

					continue
				}
				c.logger.Debug("Sensor registered.",
					sensorData.LogAttributes())
			}
			// Add sensor details to the tracker.
			if err := c.sensorTracker.Add(&sensorData); err != nil {
				c.logger.Warn("Updating sensor tracker failed.",
					sensorData.LogAttributes(),
					slog.Any("error", err),
				)
			}

			continue
		}

		c.logger.Warn("Unhandled entity recieved.",
			slog.String("entity_type", fmt.Sprintf("%T", entity)))
	}
}

// SendRequest will send the given request to the specified URL. It will handle
// marshaling the request and unmarshaling the response. It will also handle
// retrying the request with an exponential backoff if requested.
func (c *Client) SendRequest(ctx context.Context, url string, req api.Request) (api.Response, error) {
	var resp api.Response

	// Set up the api request, and the request/response bodies.
	apiReq := c.restAPI.R().SetContext(ctx)
	apiReq.SetBody(req)
	apiReq = apiReq.SetResult(&resp)

	// If request needs to be retried, retry the request on any error.
	if req.Retryable != nil && *req.Retryable {
		apiReq = apiReq.AddRetryCondition(
			func(_ *resty.Response, err error) bool {
				return err != nil
			},
		)
	}

	c.logger.
		LogAttrs(ctx, logging.LevelTrace,
			"Sending request.",
			slog.Group("request",
				slog.String("method", "POST"),
				slog.String("url", url),
				slog.Any("body", req.Data),
				slog.Time("sent_at", time.Now()),
			),
		)

	// Send the request.
	apiResp, err := apiReq.Post(url)
	// Handle different response conditions.
	switch {
	case err != nil:
		return resp, errors.Join(ErrSendRequest, err)
	case apiResp == nil:
		return resp, errors.Join(ErrSendRequest, errors.New("an unknown error occurred"))
	case apiResp.IsError():
		return resp, errors.Join(ErrSendRequest, fmt.Errorf("%s", apiResp.Status()))
	}

	c.logger.
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Group("response",
				slog.Int("statuscode", apiResp.StatusCode()),
				slog.String("status", apiResp.Status()),
				slog.String("protocol", apiResp.Proto()),
				slog.Duration("time", apiResp.Time()),
				slog.String("body", string(apiResp.Body())),
			),
		)

	return resp, nil
}

// GetHAVersion retrieves the Home Assistant version.
func (c *Client) GetHAVersion(ctx context.Context) string {
	config, err := c.getHAConfig(ctx)
	if err != nil {
		c.logger.Warn("Could not fetch Home Assistant config.",
			slog.Any("error", err))
		return "Unknown"
	}

	return config.GetVersion()
}

func (c *Client) GetSensorList() []models.UniqueID {
	return c.sensorTracker.SensorList()
}

func (c *Client) GetSensor(id models.UniqueID) (*models.Sensor, error) {
	return c.sensorTracker.Get(id)
}

func (c *Client) DisableSensor(id models.UniqueID) {
	if !c.isDisabledInReg(id) {
		c.logger.Info("Disabling sensor.",
			slog.String("id", id))
		c.sensorRegistry.SetDisabled(id, true)
	}
}

// NewClient creates a new hass client, which tracks last sensor status,
// sensor registration status and handles sending and processing requests to the
// Home Assistant REST API.
func NewClient(ctx context.Context, reg sensorRegistry) *Client {
	return &Client{
		logger:         logging.FromContext(ctx).WithGroup("hass"),
		sensorRegistry: reg,
		sensorTracker:  tracker.NewTracker(),
		restAPI: resty.New().
			SetTimeout(defaultTimeout).
			SetRetryCount(defaultRetryCount).
			SetRetryWaitTime(defaultRetryWait).
			SetRetryMaxWaitTime(defaultRetryMaxWait).
			AddRetryCondition(func(r *resty.Response, _ error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			}),
	}
}
