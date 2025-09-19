// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package agent defines the core functionality for running the agent.
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
)

var registered chan struct{}

const (
	agentConfigPrefix = "agent"
)

// Agent represents the data and methods required for running the agent.
type Agent struct {
	Config *Config
}

// Config contains the agent configuration options.
type Config struct {
	Registered bool   `toml:"registered"`
	ID         string `toml:"device_id"`
}

// New sets up a new agent instance.
func New() (*Agent, error) {
	registered = make(chan struct{})

	agent := &Agent{
		Config: &Config{
			Registered: false,
		},
	}
	// Load the server config.
	if err := config.Load(agentConfigPrefix, agent.Config); err != nil {
		return agent, fmt.Errorf("unable to load agent config: %w", err)
	}
	// Generate a unique device ID if required.
	if agent.Config.ID == "" {
		slog.Debug("Generating new device ID.")
		// Generate a new unique Device ID
		id, err := device.NewDeviceID()
		if err != nil {
			return agent, fmt.Errorf("unable to generate new device id: %w", err)
		}
		err = config.Set(map[string]any{"agent.device_id": id})
		if err != nil {
			return agent, fmt.Errorf("unable to generate new device id: %w", err)
		}
		agent.Config.ID = id
	}

	if agent.IsRegistered() {
		close(registered)
	}
	// Return the controller object.
	return agent, nil
}

// IsRegistered returns a boolean indicating whether the agent has been registered.
func (a *Agent) IsRegistered() bool {
	return a.Config.Registered
}

// Register will mark the registration status of the agent as registered.
func (a *Agent) Register() {
	err := config.Set(map[string]any{"agent.registered": true})
	if err != nil {
		slog.Error("Unable to save registration status to config.",
			slog.Any("error", err))
		return
	}
	a.Config.Registered = true
	close(registered)
}

// Run is the main loop of the agent. It will configure and run all sensor workers and process and send the data to Home
// Assistant. Run blocks and won't perform any actions until the registration status of the agent is true.
func (a *Agent) Run(ctx context.Context) error {
	for {
		select {
		case <-registered:
			slog.Debug("Agent is registered.")
			// hassClient, err := hass.NewClient(ctx)
			// godump.Dump(hassClient)
			// if err != nil {
			// 	return fmt.Errorf("unable to run agent: %w", err)
			// }
			return nil
		case <-ctx.Done():
			slog.Debug("Stopping agent.")
			return nil
		}
	}
}
