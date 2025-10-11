// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package cli contains methods for handling the command-line of the agent.
package cli

import (
	"embed"
)

// Opts are the global command-line options common across all commands.
type Opts struct {
	Path          string
	StaticContent embed.FS
}
