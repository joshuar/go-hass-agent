// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrMQTTServerRequired = errors.New("mqtt-server not specified")

type ConfigCmd struct {
	Path       string `kong:"hidden"`
	MQTTConfig `kong:"help='Set MQTT options.'"`
}

type MQTTConfig preferences.MQTT

func (r *ConfigCmd) Run(ctx *Context) error {
	r.Path = filepath.Join(xdg.ConfigHome, ctx.AppID)

	prefs, err := preferences.Load(r.Path)
	if err != nil {
		return fmt.Errorf("config: load preferences: %w", err)
	}

	r.MQTTEnabled = true
	prefs.MQTT = (*preferences.MQTT)(&r.MQTTConfig)

	err = prefs.Save()
	if err != nil {
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
