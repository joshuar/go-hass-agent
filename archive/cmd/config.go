// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

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
	ctx = slogctx.NewCtx(ctx, slog.Default())
	ctx = preferences.PathToCtx(ctx, opts.Path)

	if err := preferences.Init(ctx); err != nil {
		return fmt.Errorf("config: load preferences: %w", err)
	}

	var preferencesToSet []preferences.SetPreference

	// Enable MQTT functionality in the preferences.
	preferencesToSet = append(preferencesToSet, preferences.SetMQTTEnabled(true))
	// Set the user if it was passed in.
	if r.MQTTUser != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTUser(r.MQTTUser))
	}
	// Set the password if it was passed in.
	if r.MQTTPassword != "" {
		preferencesToSet = append(preferencesToSet, preferences.SetMQTTPassword(r.MQTTPassword))
	}
	// If no server was given, set the server to the default.
	if r.MQTTServer == "" {
		r.MQTTServer = preferences.DefaultMQTTServer
	}
	preferencesToSet = append(preferencesToSet, preferences.SetMQTTServer(r.MQTTServer))
	// If no topic prefix was passed in, set it to the default.
	if r.MQTTTopicPrefix == "" {
		r.MQTTTopicPrefix = preferences.DefaultMQTTTopicPrefix
	}
	preferencesToSet = append(preferencesToSet, preferences.SetMQTTTopicPrefix(r.MQTTTopicPrefix))

	if err := preferences.Set(preferencesToSet...); err != nil {
		return fmt.Errorf("config: save preferences: %w", err)
	}

	return nil
}
