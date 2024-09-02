// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"log/slog"
	"os"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/internal/cli"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

//nolint:tagalign
var CLI struct {
	Run       cli.RunCmd       `cmd:"" help:"Run Go Hass Agent."`
	Reset     cli.ResetCmd     `cmd:"" help:"Reset Go Hass Agent."`
	Version   cli.VersionCmd   `cmd:"" help:"Show the Go Hass Agent version."`
	Upgrade   cli.UpgradeCmd   `cmd:"" help:"Attempt to upgrade from previous version."`
	Profile   cli.ProfileFlags `help:"Enable profiling."`
	AppID     string           `name:"appid" default:"${defaultAppID}" help:"Specify a custom app id (for debugging)."`
	LogLevel  string           `name:"log-level" enum:"info,debug,trace" default:"info" help:"Set logging level."`
	Register  cli.RegisterCmd  `cmd:"" help:"Register with Home Assistant."`
	NoLogFile bool             `help:"Don't write to a log file."`
	Headless  bool             `name:"terminal" help:"Run without a GUI."`
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
	kong.Name(preferences.AppName)
	kong.Description(preferences.AppDescription)
	ctx := kong.Parse(&CLI, kong.Bind(), kong.Vars{"defaultAppID": preferences.AppID})
	// Warn if running headless without explicitly specifying so.
	if os.Getenv("DISPLAY") == "" {
		if !CLI.Headless {
			slog.Warn("DISPLAY not set, running in headless mode by default (specify --terminal to suppress this warning).")
		}

		CLI.Headless = true
	}

	err := ctx.Run(cli.CreateCtx(
		cli.RunHeadless(CLI.Headless),
		cli.WithProfileFlags(CLI.Profile),
		cli.WithAppID(CLI.AppID),
		cli.WithLogLevel(CLI.LogLevel),
		cli.WithLogFile(!CLI.NoLogFile),
	))
	if CLI.Profile != nil {
		err = errors.Join(logging.StopProfiling(logging.ProfileFlags(CLI.Profile)), err)
	}

	ctx.FatalIfErrorf(err)
}
