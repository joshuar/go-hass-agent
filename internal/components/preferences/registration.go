// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
)

var ErrRegistrationPreference = errors.New("registration preference error")

// Registration are the preferences that defines how Go Hass Agent registers
// with Home Assistant.
type Registration struct {
	Server         string `toml:"server" validate:"required,http_url"`
	Token          string `toml:"token" validate:"required"`
	IgnoreHassURLs bool   `toml:"-" json:"-" validate:"omitempty,boolean"`
	ForceRegister  bool   `toml:"-" json:"-" validate:"omitempty,boolean"`
}

// Validate will check the registration preferences are valid.
func (p *Registration) Validate() error {
	err := validation.Validate.Struct(p)
	if err != nil {
		return fmt.Errorf("%w: validation failed: %s", ErrRegistrationPreference, validation.ParseValidationErrors(err))
	}

	return nil
}

// Server returns the server that was used for registering Go Hass Agent.
func (p *preferences) Server() string {
	return p.Registration.Server
}

// Token returns the long-lived access token that was used by Go Hass Agent for
// registering with Home Assistant.
func (p *preferences) Token() string {
	return p.Registration.Token
}
