// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/config"
)

var ErrVersionCmdFailed = errors.New("version command failed")

// Version: `go-hass-agent version`.
type Version struct{}

func (r *Version) Run(_ *Opts) error {
	_, err := fmt.Fprintf(os.Stdout, "%s: %s\n", config.AppName, config.AppVersion)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrVersionCmdFailed, err)
	}
	return nil
}
