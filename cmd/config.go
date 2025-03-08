// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

var ErrMQTTServerRequired = errors.New("mqtt-server not specified")

// Config: `go-hass-agent config`.
type Config struct {
	MQTTConfig `kong:"help='Set MQTT options.'"`
}

type MQTTConfig preferences.MQTTPreferences

func (r *Config) Run(opts *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	ctx = preferences.PathToCtx(ctx, opts.Path)

	if err := preferences.Init(ctx); err != nil {
		return fmt.Errorf("config: load preferences: %w", err)
	}

	var preferencesToSet []preferences.SetPreference

	preferencesToSet = append(preferencesToSet, preferences.SetMQTTEnabled(true))

	if r.MQTTServer != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTServer(r.MQTTServer))
	}

	if r.MQTTTopicPrefix != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTTopicPrefix(r.MQTTTopicPrefix))
	}

	if r.MQTTUser != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTUser(r.MQTTUser))
	}

	if r.MQTTPassword != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTPassword(r.MQTTPassword))
	}

	if err := preferences.Set(preferencesToSet...); err != nil {
		return fmt.Errorf("config: save preferences: %w", err)
	}

	return nil
}
