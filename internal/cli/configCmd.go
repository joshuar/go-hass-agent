// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrMQTTServerRequired = errors.New("mqtt-server not specified")

// ConfigCmd: `go-hass-agent config`.
type ConfigCmd struct {
	MQTTConfig `kong:"help='Set MQTT options.'"`
}

type MQTTConfig preferences.MQTT

func (r *ConfigCmd) Run(_ *CmdOpts) error {
	if err := preferences.Load(); err != nil {
		return fmt.Errorf("config: load preferences: %w", err)
	}

	r.MQTTEnabled = true
	if err := preferences.SetMQTTPreferences((*preferences.MQTT)(&r.MQTTConfig)); err != nil {
		return fmt.Errorf("config: save preferences: %w", err)
	}

	if err := preferences.Save(); err != nil {
		return fmt.Errorf("config: save preferences: %w", err)
	}

	return nil
}

func (r *MQTTConfig) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrValidationFailed, parseValidationErrors(err))
	}

	if r.MQTTServer == "" {
		return ErrMQTTServerRequired
	}

	return nil
}
