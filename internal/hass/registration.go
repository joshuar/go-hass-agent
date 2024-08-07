// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:errname // structs are dual-purpose response and error
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

type registrationResponse struct {
	Details *preferences.Hass
	*APIError
}

func (r *registrationResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &r.Details)
	if err != nil {
		return fmt.Errorf("failed to parse registration response: %w", err)
	}

	return nil
}

func (r *registrationResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, r.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (r *registrationResponse) Error() string {
	return r.APIError.Error()
}

func newRegistrationResponse() *registrationResponse {
	return &registrationResponse{}
}

func RegisterDevice(ctx context.Context, device *preferences.Device, registration *preferences.Registration) (*preferences.Hass, error) {
	// Validate provided registration details.
	if err := registration.Validate(); err != nil {
		// if !validRegistrationSetting("server", input.Server) || !validRegistrationSetting("token", token) {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	// Generate request and response objects.
	req := newRegistrationRequest(device, registration.Token)
	resp := newRegistrationResponse()

	// Connect to specified Home Assistant server.
	client := NewDefaultHTTPClient(registration.Server).SetTimeout(time.Minute)

	// Register the device against the registration endpoint.
	if err := ExecuteRequest(ctx, client, RegistrationPath, req, resp); err != nil {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	return resp.Details, nil
}
