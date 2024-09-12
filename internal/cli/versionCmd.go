// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type VersionCmd struct{}

func (r *VersionCmd) Run(_ *CmdOpts) error {
	fmt.Fprintf(os.Stdout, "%s: %s\n", preferences.AppName, preferences.AppVersion)

	return nil
}
