// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/validation"
)

type Registration struct {
	Server         string `toml:"server" validate:"required,http_url"`
	Token          string `toml:"token" validate:"required"`
	IgnoreHassURLs bool   `toml:"-" json:"-" validate:"omitempty,boolean"`
}

func (p *Registration) Validate() error {
	err := validation.Validate.Struct(p)
	if err != nil {
		return fmt.Errorf("validation failed: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

func (p *Preferences) Server() string {
	return p.Registration.Server
}

func (p *Preferences) Token() string {
	return p.Registration.Token
}
