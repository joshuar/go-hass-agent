// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
)

// Config represents the options for the `config` command.
type Config struct {
	mqtt.Config
}

// AfterApply applies MQTT config values that are not exposed via command-line options, as necessary.
func (c *Config) AfterApply() error {
	if c.MQTTTopicPrefix == "" {
		c.MQTTTopicPrefix = mqtt.DefaultTopicPrefix
	}
	if c.MQTTServer != "" {
		c.MQTTEnabled = true
	}
	return nil
}

// Run processes the config command.
func (c *Config) Run(_ *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())

	// Validate config options.
	valid, err := c.Valid()
	if !valid || err != nil {
		return fmt.Errorf("unable to validate preferences: %w", err)
	}

	err = config.Save(mqtt.ConfigPrefix, c.Config)
	if err != nil {
		return fmt.Errorf("unable to save preferences: %w", err)
	}
	slogctx.FromCtx(ctx).Info("Preferences saved!")

	return nil
}
