// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package agent defines the core functionality for running the agent.
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/config"
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
	Registered bool `toml:"registered"`
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
