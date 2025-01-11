// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	RegistrationPath = "/api/mobile_app/registrations"
)

var ErrInternalValidationFailed = errors.New("internal validation error")

type registrationRequest struct {
	*preferences.Device
	Token string `json:"-"`
}

func (r *registrationRequest) Auth() string {
	return r.Token
}

func (r *registrationRequest) RequestBody() any {
	return r.Device
}

// revive:disable:unused-receiver
func (r *registrationRequest) Retry() bool {
	return true
}

func newRegistrationRequest(thisDevice *preferences.Device, token string) *registrationRequest {
	return &registrationRequest{
		Device: thisDevice,
		Token:  token,
	}
}

func RegisterDevice(ctx context.Context, registration *preferences.Registration) (*preferences.Hass, error) {
	// Validate provided registration details.
	if err := registration.Validate(); err != nil {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	registrationURL := registration.Server + RegistrationPath
	dev := preferences.NewDevice()

	// Register the device against the registration endpoint.
	registrationStatus, err := api.Send[preferences.Hass](ctx, registrationURL, newRegistrationRequest(dev, registration.Token))
	if err != nil {
		return nil, fmt.Errorf("could not register device: %w", err)
	}

	return &registrationStatus, nil
}
