// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/config"
)

// Version is the command-line option for showing the agent version.
type Version struct{}

// Run will run the version command.
func (r *Version) Run(_ *Opts) error {
	_, err := fmt.Fprintf(os.Stdout, "%s: %s\n", config.AppName, config.AppVersion)
	if err != nil {
		return fmt.Errorf("unable to show version: %w", err)
	}
	return nil
}
