// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"log/slog"
	"os"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/cli"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/logging"
)

//go:embed all:web/content
var content embed.FS

// CLI contains all of the commands and common options for running Go Hass
// Agent.
var CLI struct {
	logging.Options

	Run cli.Run `cmd:"" help:"Run Go Hass Agent."`
	// Reset        cmd.Reset            `cmd:"" help:"Reset Go Hass Agent."`
	Version cli.Version `cmd:"" help:"Show the Go Hass Agent version."`
	// Upgrade      cmd.Upgrade          `cmd:"" help:"Attempt to upgrade from previous version."`
	ProfileFlags logging.ProfileFlags `name:"profile" help:"Set profiling flags."`
	Config       cli.Config           `cmd:"" help:"Configure Go Hass Agent."`
	Register     cli.Register         `cmd:"" help:"Register with Home Assistant."`
	Registry     cli.RegistryCmd      `cmd:"" help:"Registry actions"`
	Path         string               `name:"path" default:"${defaultPath}" help:"Specify a custom path to store preferences/logs/data (for debugging)."`
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
	kong.Name(config.AppName)
	kong.Description(config.AppDescription)
	// Parse the command-line.
	cmdCtx := kong.Parse(&CLI, kong.Bind(), kong.Vars{"defaultPath": config.GetPath()})
	config.SetPath(CLI.Path)
	// Set up the logger.
	logging.New(logging.Options{LogLevel: CLI.LogLevel, NoLogFile: CLI.NoLogFile})
	// Enable profiling if requested.
	if CLI.ProfileFlags != nil {
		if err := logging.StartProfiling(CLI.ProfileFlags); err != nil {
			slog.Warn("Problem starting profiling.",
				slog.Any("error", err))
		}
	}
	// Initialise the config.
	err := config.Init()
	if err != nil {
		slog.Error("Unable to start.",
			slog.Any("error", err))
		os.Exit(-1)
	}
	// Generate device details if required.
	if !config.Exists(device.ConfigPrefix) {
		err := device.NewConfig()
		if err != nil {
			slog.Error("Unable to start.",
				slog.Any("error", err))
			os.Exit(-1)
		}
	}
	// Run the requested command with the provided options.
	if err := cmdCtx.Run(&cli.Opts{Path: CLI.Path, StaticContent: content}); err != nil {
		slog.Error("Command failed.",
			slog.String("command", cmdCtx.Command()),
			slog.Any("error", err))
	}
	// If profiling was enabled, clean up.
	if CLI.ProfileFlags != nil {
		if err := logging.StopProfiling(CLI.ProfileFlags); err != nil {
			slog.Error("Problem stopping profiling.",
				slog.Any("error", err))
		}
	}
}
