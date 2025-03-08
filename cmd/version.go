// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cmd

import (
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

// Version: `go-hass-agent version`.
type Version struct{}

func (r *Version) Run(_ *Opts) error {
	fmt.Fprintf(os.Stdout, "%s: %s\n", preferences.AppName, preferences.AppVersion())

	return nil
}
