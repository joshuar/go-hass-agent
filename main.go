// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"os"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/internal/cli"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// CLI contains all of the commands and common options for running Go Hass
// Agent.
var CLI struct {
	Run          cli.RunCmd           `cmd:"" help:"Run Go Hass Agent."`
	Reset        cli.ResetCmd         `cmd:"" help:"Reset Go Hass Agent."`
	Version      cli.VersionCmd       `cmd:"" help:"Show the Go Hass Agent version."`
	Upgrade      cli.UpgradeCmd       `cmd:"" help:"Attempt to upgrade from previous version."`
	ProfileFlags logging.ProfileFlags `name:"profile" help:"Set profiling flags."`
	Headless     *cli.HeadlessFlag    `name:"terminal" help:"Run without a GUI." default:"false"`
	AppID        string               `name:"appid" default:"${defaultAppID}" help:"Specify a custom app id (for debugging)."`
	Config       cli.ConfigCmd        `cmd:"" help:"Configure Go Hass Agent."`
	Register     cli.RegisterCmd      `cmd:"" help:"Register with Home Assistant."`
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
	ctx := kong.Parse(&CLI, kong.Bind(), kong.Vars{"defaultAppID": preferences.AppID()})
	// Set up the logger.
	logger := logging.New(CLI.AppID, logging.Options{LogLevel: CLI.LogLevel, NoLogFile: CLI.NoLogFile})
	// Enable profiling if requested.
	if CLI.ProfileFlags != nil {
		if err := logging.StartProfiling(CLI.ProfileFlags); err != nil {
			logger.Warn("Problem starting profiling.",
				slog.Any("error", err))
		}
	}
	// Run the requested command with the provided options.
	if err := ctx.Run(cli.AddOptions(
		cli.RunHeadless(bool(*CLI.Headless)),
		cli.WithAppID(CLI.AppID),
		cli.WithLogger(logger),
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
