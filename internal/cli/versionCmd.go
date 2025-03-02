// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

// VersionCmd: `go-hass-agent version`.
type VersionCmd struct{}

func (r *VersionCmd) Run(_ *CmdOpts) error {
	fmt.Fprintf(os.Stdout, "%s: %s\n", preferences.AppName, preferences.AppVersion())

	return nil
}
