// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/reugn/go-quartz/job"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/hass/event"
	"github.com/joshuar/go-hass-agent/hass/location"
	"github.com/joshuar/go-hass-agent/hass/registry"
	"github.com/joshuar/go-hass-agent/hass/sensor"
	"github.com/joshuar/go-hass-agent/hass/tracker"
	"github.com/joshuar/go-hass-agent/id"
	"github.com/joshuar/go-hass-agent/logging"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	hassConfigPrefix = "hass"
)

type agent interface {
	IsRegistered() bool
}

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
	config         *Config
}

// ErrSendRequest indicates an error occurred when sending a request to Home Assistant.
var ErrSendRequest = errors.New("send request failed")

// NewClient creates a new hass client, which tracks last sensor status,
// sensor registration status and handles sending and processing requests to the
// Home Assistant REST API.
func NewClient(ctx context.Context, agent agent) (*Client, error) {
	var hasscfg Config
	// Load the hass config.
	if err := config.Load(hassConfigPrefix, &hasscfg); err != nil {
		return nil, fmt.Errorf("unable to load hass config: %w", err)
	}
	// Load the registry.
	reg, err := registry.Load(config.GetPath())
	if err != nil {
		return nil, fmt.Errorf("unable to create hass client: %w", err)
	}
	// Create the client.
	client := &Client{
		logger:         slogctx.FromCtx(ctx).WithGroup("hass"),
		sensorRegistry: reg,
		sensorTracker:  tracker.NewTracker(),
		config:         &hasscfg,
		restAPI:        api.NewClient(),
	}
	// Run the job one-time initially to get the config.
	if agent.IsRegistered() {
		updated, err := client.UpdateConfig(ctx)
		if !updated || err != nil {
			return nil, fmt.Errorf("could not create client: %w", err)
		}
		// Schedule a job to get the Home Assistant on a regular interval.
		if err := client.scheduleConfigUpdates(); err != nil {
			return nil, fmt.Errorf("could not create client: %w", err)
		}
	}

	return client, nil
}

func (c *Client) RestAPIURL() string {
	return c.config.APIURL
}

// UpdateConfig will fetch and store the Home Assistant config via the Home Assistant REST API.
func (c *Client) UpdateConfig(ctx context.Context) (bool, error) {
	resp, err := c.SendRequest(ctx, c.RestAPIURL(), api.Request{Type: api.GetConfig})
	if err != nil {
		return false, fmt.Errorf("could not update config: %w", err)
	}

	configResp, err := resp.AsConfigResponse()
	if err != nil {
		return false, fmt.Errorf("could not update config: %w", err)
	}

	c.config.Update(&configResp)

	return true, nil
}

// EntityHandler takes incoming Entity objects via the passed in channel and
// runs the appropriate handler for the Entity type.
//
//nolint:gocognit,gocyclo,funlen
func (c *Client) EntityHandler(ctx context.Context, entityCh <-chan models.Entity) {
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

		//nolint:nestif
		if sensorData, err := entity.AsSensor(); err == nil {
			// Send sensor details.
			if c.sensorRegistry.IsRegistered(sensorData.UniqueID) {
				// Ignore updates for disabled sensors.
				if c.isDisabled(ctx, sensorData) {
					continue
				}
				// Otherwise, send an update.
				if err := sensor.UpdateHandler(ctx, c, sensorData); err != nil {
					c.logger.Warn("Could not update sensor.",
						sensorData.LogAttributes(),
						slog.Any("error", err))

					continue
				}

				c.logger.Log(ctx, logging.LevelTrace, "Sensor updated.",
					sensorData.LogAttributes())
			} else {
				// Otherwise, send a registration request.
				success, err := sensor.RegistrationHandler(ctx, c, sensorData)

				switch {
				case err != nil:
					c.logger.Warn("Send sensor registration failed.",
						sensorData.LogAttributes(),
						slog.Any("error", err))
				case !success:
					c.logger.Warn("Sensor not registered.",
						sensorData.LogAttributes())
				default:
					if err := c.sensorRegistry.SetRegistered(sensorData.UniqueID, true); err != nil {
						c.logger.Warn("Could not set local registration status.",
							slog.Any("error", err))
						continue
					}

					c.logger.Debug("Sensor registered.",
						sensorData.LogAttributes())
				}
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

		c.logger.Warn("Unhandled entity received.",
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
	if req.Retryable {
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
				// slog.Any("body", req.Data),
				slog.Time("sent_at", time.Now()),
			),
		)

	// Send the request.
	apiResp, err := apiReq.Post(url)
	// Handle different response conditions.
	switch {
	case err != nil:
		return resp, fmt.Errorf("%w: %w", ErrSendRequest, err)
	case apiResp == nil:
		return resp, fmt.Errorf("%w: an unknown error occurred", ErrSendRequest)
	case apiResp.IsError():
		return resp, fmt.Errorf("%w: %s", ErrSendRequest, apiResp.Status())
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
func (c *Client) GetHAVersion() string {
	return c.config.GetVersion()
}

func (c *Client) GetSensorList() []models.UniqueID {
	return c.sensorTracker.SensorList()
}

func (c *Client) GetSensor(id models.UniqueID) (*models.Sensor, error) {
	return c.sensorTracker.Get(id) //nolint:wrapcheck
}

func (c *Client) DisableSensor(id models.UniqueID) {
	if !c.isDisabledInReg(id) {
		c.logger.Debug("Disabling sensor.",
			slog.String("id", id))
		if err := c.sensorRegistry.SetDisabled(id, true); err != nil {
			c.logger.Warn("Could not disable sensor.", slog.Any("error", err))
		}
	}
}

func (c *Client) RegisterSensor(id models.UniqueID) error {
	if err := c.sensorRegistry.SetRegistered(id, true); err != nil {
		return fmt.Errorf("could not register sensor: %w", err)
	}
	return nil
}

// Reset performs a reset of the client. It will remove existing registry data.
func Reset() error {
	err := registry.Reset(config.GetPath())
	if err != nil {
		return fmt.Errorf("unable to reset client: %w", err)
	}
	return nil
}

// isDisabled handles processing a sensor that is disabled. For a sensor that is
// disabled, we need to make an additional check against Home Assistant to see
// if the sensor has been re-enabled, and update our local registry before
// continuing.
func (c *Client) isDisabled(ctx context.Context, details models.Sensor) bool {
	regDisabled := c.isDisabledInReg(details.UniqueID)
	haDisabled, haConfigErr := c.isDisabledInHA(details.UniqueID)

	switch {
	case regDisabled && (haConfigErr == nil && !haDisabled):
		c.logger.Debug("Sensor re-enabled in Home Assistant, Re-enabling in local registry and sending updates.",
			details.LogAttributes())
		if err := c.sensorRegistry.SetDisabled(details.UniqueID, false); err != nil {
			slogctx.FromCtx(ctx).Warn("Could not update sensor state in registry.",
				details.LogAttributes())
		}
		return false
	case !regDisabled && (haConfigErr == nil && haDisabled):
		c.logger.Debug("Sensor has been disabled in Home Assistant, Disabling in local registry and not sending updates.",
			details.LogAttributes())
		if err := c.sensorRegistry.SetDisabled(details.UniqueID, true); err != nil {
			slogctx.FromCtx(ctx).Warn("Could not update sensor state in registry.",
				details.LogAttributes())
		}
		return true
	case regDisabled && (haConfigErr == nil && haDisabled):
		c.logger.Debug("Sensor is disabled, not sending updates.",
			details.LogAttributes())

		return true
	}

	return false
}

// isDisabledInReg returns the disabled state of the sensor from the local
// registry.
func (c *Client) isDisabledInReg(id models.UniqueID) bool {
	return c.sensorRegistry.IsDisabled(id)
}

// isDisabledInHA returns the disabled state of the sensor from Home Assistant.
func (c *Client) isDisabledInHA(id models.UniqueID) (bool, error) {
	status, err := c.config.IsEntityDisabled(id)
	if err != nil {
		return false, err
	}

	return status, nil
}

func (c *Client) scheduleConfigUpdates() error {
	if scheduler.Manager == nil {
		c.logger.Debug("No scheduler active, not scheduling fetch config updates.")
		return nil
	}
	getConfigJob := job.NewFunctionJobWithDesc(c.UpdateConfig, "Fetch Home Assistant Configuration.")
	err := scheduler.Manager.ScheduleJob(id.HassJob, getConfigJob, quartz.NewSimpleTrigger(30*time.Second))
	if err != nil {
		return fmt.Errorf("could not schedule config updates: %w", err)
	}
	return nil
}
