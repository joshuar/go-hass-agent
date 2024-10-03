// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//go:generate go run github.com/matryer/moq -out client_mocks_test.go . PostRequest Registry
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	DefaultTimeout = 30 * time.Second
)

var tracker = sensor.NewTracker()

var (
	ErrGetConfigFailed   = errors.New("could not fetch Home Assistant config")
	ErrGenRequestFailed  = errors.New("unable to generate request for sensor")
	ErrSendRequestFailed = errors.New("could not send sensor request to Home Assistant")

	ErrStateUpdateUnknown = errors.New("unknown sensor update response")
	ErrStateUpdateFailed  = errors.New("state update failed")
	ErrRegDisableFailed   = errors.New("failed to disable sensor in registry")
	ErrRegAddFailed       = errors.New("failed to set registered status for sensor in registry")
	ErrTrkUpdateFailed    = errors.New("failed to update sensor state in tracker")
	ErrRegistrationFailed = errors.New("sensor registration failed")

	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidClient     = errors.New("invalid client")
	ErrResponseMalformed = errors.New("malformed response")
	ErrUnknown           = errors.New("unknown error occurred")

	ErrInvalidSensor = errors.New("invalid sensor")

	defaultRetry = func(r *resty.Response, _ error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	}
)

// Validate is a request that supports validation of its values.
type Validate interface {
	Validate() error
}

// GetRequest is a HTTP GET request.
type GetRequest any

// PostRequest is a HTTP POST request with the request body provided by Body().
type PostRequest interface {
	RequestBody() json.RawMessage
}

// Authenticated represents a request that requires passing an authentication
// header with the value returned by Auth().
type Authenticated interface {
	Auth() string
}

// Encrypted represents a request that should be encrypted with the secret
// provided by Secret().
type Encrypted interface {
	Secret() string
}

type Registry interface {
	SetDisabled(id string, state bool) error
	SetRegistered(id string, state bool) error
	IsDisabled(id string) bool
	IsRegistered(id string) bool
}

type Client struct {
	endpoint *resty.Client
	registry Registry
}

func NewClient(ctx context.Context) (*Client, error) {
	var err error

	reg, err := registry.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start registry: %w", err)
	}

	client := &Client{
		registry: reg,
	}

	return client, nil
}

func (c *Client) Endpoint(url string, timeout time.Duration) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	c.endpoint = resty.New().
		SetTimeout(timeout).
		AddRetryCondition(defaultRetry).
		SetBaseURL(url)
}

func (c *Client) HassVersion(ctx context.Context) string {
	config, err := send[Config](ctx, c, &configRequest{})
	if err != nil {
		logging.FromContext(ctx).
			Debug("Could not fetch Home Assistant config.",
				slog.Any("error", err))

		return "Unknown"
	}

	return config.Version
}

func (c *Client) ProcessSensor(ctx context.Context, details sensor.Entity) error {
	if c.isDisabled(ctx, details) {
		logging.FromContext(ctx).
			Debug("Not sending request for disabled sensor.",
				sensorLogAttrs(details))

		return nil
	}

	if _, ok := details.Value.(*LocationRequest); ok {
		// LocationRequest:
		return c.handleLocationUpdate(ctx, details)
	}

	if c.registry.IsRegistered(details.ID) {
		// Sensor Update (existing sensor).
		return c.handleSensorUpdate(ctx, details)
	}
	// Sensor Registration (new sensor).
	return c.handleRegistration(ctx, details)
}

func (c *Client) handleLocationUpdate(ctx context.Context, details sensor.Entity) error {
	// req, err := sensor.NewLocationUpdateRequest(details)
	req, err := newEntityRequest(requestTypeLocation, details)
	if err != nil {
		return fmt.Errorf("unable to handle location update: %w", err)
	}

	resp, err := send[locationResponse](ctx, c, req)
	if err != nil {
		return fmt.Errorf("failed to send location update request: %w", err)
	}

	if err := resp.updated(); err != nil { //nolint:staticcheck
		return fmt.Errorf("location update failed: %w", err)
	}

	return nil
}

func (c *Client) handleSensorUpdate(ctx context.Context, details sensor.Entity) error {
	// req, err := sensor.NewUpdateRequest(details)
	req, err := newEntityRequest(requestTypeUpdate, details)
	if err != nil {
		return fmt.Errorf("unable to handle sensor update: %w", err)
	}

	response, err := send[stateUpdateResponse](ctx, c, req)
	if err != nil {
		return fmt.Errorf("failed to send sensor update request for %s: %w", details.ID, err)
	}

	if response == nil {
		return ErrStateUpdateUnknown
	}

	// At this point, the sensor update was successful. Any errors are really
	// warnings and non-critical.
	var warnings error

	for id, update := range response {
		success, err := update.success()
		if !success {
			// The update failed.
			warnings = errors.Join(warnings, err)
		}

		// If HA reports the sensor as disabled, update the registry.
		if c.registry.IsDisabled(id) != update.disabled() {
			logging.FromContext(ctx).
				Info("Sensor is disabled in Home Assistant. Setting disabled in local registry.",
					sensorLogAttrs(details))

			if err := c.registry.SetDisabled(id, update.disabled()); err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrRegDisableFailed, err))
			}
		}

		// Add the sensor update to the tracker.
		if err := tracker.Add(&details); err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrTrkUpdateFailed, err))
		}
	}

	if warnings != nil {
		logging.FromContext(ctx).
			Debug("Sensor updated with warnings.",
				sensorLogAttrs(details),
				slog.Any("warnings", warnings))
	} else {
		logging.FromContext(ctx).
			Debug("Sensor updated.",
				sensorLogAttrs(details))
	}
	// Return success status and any warnings.
	return warnings
}

func (c *Client) handleRegistration(ctx context.Context, details sensor.Entity) error {
	req, err := newEntityRequest(requestTypeRegister, details)
	if err != nil {
		return fmt.Errorf("unable to handle sensor update: %w", err)
	}

	response, err := send[registrationResponse](ctx, c, req)
	if err != nil {
		return fmt.Errorf("failed to send sensor registration request for %s: %w", details.ID, err)
	}

	// If the registration failed, log a warning.
	success, err := response.registered()
	if !success {
		return errors.Join(ErrRegistrationFailed, err)
	}

	// At this point, the sensor registration was successful. Any errors are really
	// warnings and non-critical.
	var warnings error

	// Set the sensor as registered in the registry.
	err = c.registry.SetRegistered(details.ID, true)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrRegAddFailed, err))
	}
	// Update the sensor state in the tracker.
	if err := tracker.Add(&details); err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrTrkUpdateFailed, err))
	}

	// Return success status and any warnings.
	return warnings
}

// isDisabled handles processing a sensor that is disabled. For a sensor that is
// disabled, we need to make an additional check against Home Assistant to see
// if the sensor has been re-enabled, and update our local registry before
// continuing.
func (c *Client) isDisabled(ctx context.Context, details sensor.Entity) bool {
	// If it is not disabled in the local registry, immediately return false.
	if !c.isDisabledInReg(details.ID) {
		return false
	}
	// Else, get the disabled state from Home Assistant
	disabledInHA := c.isDisabledInHA(ctx, details)

	// If sensor is no longer disabled in Home Assistant, update the local
	// registry and return false.
	if !disabledInHA {
		slog.Info("Sensor re-enabled in Home Assistant, Re-enabling in local registry and sending updates.", sensorLogAttrs(details))

		if err := c.registry.SetDisabled(details.ID, false); err != nil {
			slog.Error("Could not re-enable sensor.",
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
func (c *Client) isDisabledInReg(id string) bool {
	return c.registry.IsDisabled(id)
}

// isDisabledInHA returns the disabled state of the sensor from Home Assistant.
func (c *Client) isDisabledInHA(ctx context.Context, details sensor.Entity) bool {
	config, err := send[Config](ctx, c, &configRequest{})
	if err != nil {
		logging.FromContext(ctx).
			Debug("Could not fetch Home Assistant config. Assuming sensor is still disabled.",
				sensorLogAttrs(details),
				slog.Any("error", err))

		return true
	}

	status, err := config.IsEntityDisabled(details.ID)
	if err != nil {
		logging.FromContext(ctx).
			Debug("Could not determine sensor disabled status in Home Assistant config. Assuming sensor is still disabled.",
				sensorLogAttrs(details),
				slog.Any("error", err))

		return true
	}

	return status
}

func send[T any](ctx context.Context, client *Client, requestDetails any) (T, error) {
	var (
		response    T
		responseErr apiError
		responseObj *resty.Response
	)

	if client.endpoint == nil {
		return response, ErrInvalidClient
	}

	// If the request supports validation, make sure it is valid.
	if a, ok := requestDetails.(Validate); ok {
		if err := a.Validate(); err != nil {
			return response, fmt.Errorf("validation failed: %w", err)
		}
	}

	requestObj := client.endpoint.R().SetContext(ctx)
	requestObj = requestObj.SetError(&responseErr)
	requestObj = requestObj.SetResult(&response)

	// If the request is authenticated, set the auth header with the token.
	if a, ok := requestDetails.(Authenticated); ok {
		requestObj = requestObj.SetAuthToken(a.Auth())
	}

	switch req := requestDetails.(type) {
	case PostRequest:
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "POST"),
				slog.String("body", string(req.RequestBody())),
				slog.Time("sent_at", time.Now()))

		responseObj, _ = requestObj.SetBody(req.RequestBody()).Post("") //nolint:errcheck // error is checked with responseObj.IsError()
	case GetRequest:
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "GET"),
				slog.Time("sent_at", time.Now()))

		responseObj, _ = requestObj.Get("") //nolint:errcheck // error is checked with responseObj.IsError()
	}

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Int("statuscode", responseObj.StatusCode()),
			slog.String("status", responseObj.Status()),
			slog.String("protocol", responseObj.Proto()),
			slog.Duration("time", responseObj.Time()),
			slog.String("body", string(responseObj.Body())))

	if responseObj.IsError() {
		return response, &apiError{Code: responseObj.StatusCode(), Message: responseObj.Status()}
	}

	return response, nil
}

func GetSensor(id string) (*sensor.Entity, error) {
	details, err := tracker.Get(id)
	if err != nil {
		return nil, fmt.Errorf("could not get sensor details: %w", err)
	}

	return details, nil
}

func SensorList() []string {
	return tracker.SensorList()
}

// sensorLogAttrs is a convienience function that returns some slog attributes
// for priting sensor details in the log.
func sensorLogAttrs(details sensor.Entity) slog.Attr {
	return slog.Group("sensor",
		slog.String("name", details.Name),
		slog.String("id", details.ID),
		slog.Any("state", details.Value),
		slog.String("units", details.Units))
}
