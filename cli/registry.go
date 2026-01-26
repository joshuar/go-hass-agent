// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/hass/registry"
)

// Run is the command-line option for running the agent.
type RegistryCmd struct {
	List ListRegistryCmd `cmd:"" help:"List local registry."`
}

type ListRegistryCmd struct{}

// Run starts the agent.
func (r *ListRegistryCmd) Run() error {
	reg, err := registry.Load(config.GetPath())
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	reg.List()

	return nil
}
