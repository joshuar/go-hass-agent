// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	RegistrationPath = "/api/mobile_app/registrations"
	WebsocketPath    = "/api/websocket"
	WebHookPath      = "/api/webhook/"
)

var ErrInternalValidationFailed = errors.New("internal validation error")

type registrationRequest struct {
	*preferences.Device
	Token string `json:"-"`
}

func (r *registrationRequest) Auth() string {
	return r.Token
}

func (r *registrationRequest) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return data
}

func newRegistrationRequest(device *preferences.Device, token string) *registrationRequest {
	return &registrationRequest{
		Device: device,
		Token:  token,
	}
}

func RegisterDevice(ctx context.Context, device *preferences.Device, registration *preferences.Registration) (*preferences.Hass, error) {
	// Validate provided registration details.
	if err := registration.Validate(); err != nil {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	// Create a new client connection to Home Assistant at the registration path.
	client, err := NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start hass client: %w", err)
	}

	client.Endpoint(registration.Server+RegistrationPath, time.Minute)

	// Register the device against the registration endpoint.
	registrationStatus, err := send[preferences.Hass](ctx, client, newRegistrationRequest(device, registration.Token))
	if err != nil {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	return &registrationStatus, nil
}
