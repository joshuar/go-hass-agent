// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/cmd"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

// CLI contains all of the commands and common options for running Go Hass
// Agent.
//
//nolint:lll
var CLI struct {
	Run          cmd.Run              `cmd:"" help:"Run Go Hass Agent."`
	Reset        cmd.Reset            `cmd:"" help:"Reset Go Hass Agent."`
	Version      cmd.Version          `cmd:"" help:"Show the Go Hass Agent version."`
	Upgrade      cmd.Upgrade          `cmd:"" help:"Attempt to upgrade from previous version."`
	ProfileFlags logging.ProfileFlags `name:"profile" help:"Set profiling flags."`
	Headless     *cmd.HeadlessFlag    `name:"terminal" help:"Run without a GUI." default:"false"`
	Config       cmd.Config           `cmd:"" help:"Configure Go Hass Agent."`
	Register     cmd.Register         `cmd:"" help:"Register with Home Assistant."`
	Path         string               `name:"path" default:"${defaultPath}" help:"Specify a custom path to store preferences/logs/data (for debugging)."`

	logging.Options
}

func init() {
	// Following is copied from https://git.kernel.org/pub/scm/libs/libcap/libcap.git/tree/goapps/web/web.go
	// ensureNotEUID aborts the program if it is running setuid something,
	// or being invoked by root.
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	gid := syscall.Getgid()

	if uid != euid || gid != egid || uid == 0 {
		slog.Error("go-hass-agent should not be run with additional privileges or as root.")
		os.Exit(-1)
	}
}

func main() {
	// Set some string.
	kong.Name(preferences.AppName)
	kong.Description(preferences.AppDescription)
	// Parse the command-line.
	ctx := kong.Parse(&CLI, kong.Bind(), kong.Vars{"defaultPath": filepath.Join(xdg.ConfigHome, preferences.DefaultAppID)})
	// Set up the logger.
	logger := logging.New(logging.Options{LogLevel: CLI.LogLevel, NoLogFile: CLI.NoLogFile, Path: CLI.Path})
	// Enable profiling if requested.
	if CLI.ProfileFlags != nil {
		if err := logging.StartProfiling(CLI.ProfileFlags); err != nil {
			logger.Warn("Problem starting profiling.",
				slog.Any("error", err))
		}
	}
	// Run the requested command with the provided options.
	if err := ctx.Run(cmd.AddOptions(
		cmd.RunHeadless(bool(*CLI.Headless)),
		cmd.WithLogger(logger),
		cmd.WithPath(CLI.Path),
	)); err != nil {
		logger.Error("Command failed.",
			slog.String("command", ctx.Command()),
			slog.Any("error", err))
	}
	// If profiling was enabled, clean up.
	if CLI.ProfileFlags != nil {
		if err := logging.StopProfiling(CLI.ProfileFlags); err != nil {
			logger.Error("Problem stopping profiling.",
				slog.Any("error", err))
		}
	}
}
