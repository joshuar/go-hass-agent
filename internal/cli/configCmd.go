// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"

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

	preferences.SetPreferences(
		preferences.SetMQTTEnabled(true),
	)

	if r.MQTTServer != "" {
		preferences.SetPreferences(
			preferences.SetMQTTServer(r.MQTTServer),
		)
	}

	if r.MQTTTopicPrefix != "" {
		preferences.SetPreferences(

			preferences.SetMQTTTopicPrefix(r.MQTTTopicPrefix),
		)
	}

	if r.MQTTUser != "" {
		preferences.SetPreferences(

			preferences.SetMQTTUser(r.MQTTUser),
		)
	}

	if r.MQTTPassword != "" {
		preferences.SetPreferences(
			preferences.SetMQTTPassword(r.MQTTPassword),
		)
	}

	spew.Dump(preferences.GetMQTTPreferences())

	if err := preferences.Save(); err != nil {
		return fmt.Errorf("config: save preferences: %w", err)
	}

	return nil
}
