// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package cli contains methods for handling the command-line of the agent.
package cli

import (
	"embed"
	"log/slog"
	"os"
)

// Opts are the global command-line options common across all commands.
type Opts struct {
	Path          string
	StaticContent embed.FS
}

// HeadlessFlag represents whether the agent is running headless or not.
type HeadlessFlag bool

func (f *HeadlessFlag) AfterApply() error {
	if os.Getenv("DISPLAY") == "" && !*f {
		slog.Warn("DISPLAY not set, running in headless mode by default (specify --terminal to suppress this warning).")

		*f = true
	}

	return nil
}
