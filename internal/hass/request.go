// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run github.com/matryer/moq -out request_mocks_test.go . PostRequest
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	requestTypeRegister = "register_sensor"
	requestTypeUpdate   = "update_sensor_states"
	requestTypeLocation = "update_location"
	requestTypeEvent    = "fire_event"
)

var (
	ErrNotLocation    = errors.New("sensor details do not represent a location update")
	ErrUnknownDetails = errors.New("unknown sensor details")
)

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

type Validator interface {
	Validate() error
}

// LocationRequest represents the location information that can be sent to HA to
// update the location of the agent. This is exposed so that device code can
// create location requests directly, as Home Assistant handles these
// differently from other sensors.
type LocationRequest struct {
	Gps              []float64 `json:"gps" validate:"required"`
	GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
	Battery          int       `json:"battery,omitempty"`
	Speed            int       `json:"speed,omitempty"`
	Altitude         int       `json:"altitude,omitempty"`
	Course           int       `json:"course,omitempty"`
	VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
}

type request struct {
	Data        any    `json:"data" validate:"required"`
	RequestType string `json:"type" validate:"required,oneof=register_sensor update_sensor_states update_location fire_event"`
}

func (r *request) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		return fmt.Errorf("%T is invalid: %s", r.Data, parseValidationErrors(err))
	}

	return nil
}

func (r *request) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return json.RawMessage(data)
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
