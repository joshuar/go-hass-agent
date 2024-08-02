// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type profileFlags logging.ProfileFlags

func (d profileFlags) AfterApply() error {
	err := logging.StartProfiling(logging.ProfileFlags(d))
	if err != nil {
		return fmt.Errorf("could not start profiling: %w", err)
	}

	return nil
}

type Context struct {
	Profile   profileFlags
	AppID     string
	LogLevel  string
	Headless  bool
	NoLogFile bool
}

type ResetCmd struct{}

func (r *ResetCmd) Help() string {
	return `
Reset will unregister go-hass-agent from MQTT (if in use), delete the
configuration directory and remove the log file. Use this prior to calling the
register command to start fresh.
`
}

func (r *ResetCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger := logging.New(ctx.LogLevel, ctx.NoLogFile)
	agentCtx = logging.ToContext(agentCtx, logger)

	var errs error

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to run reset command: %w", err))
	}

	// Reset agent.
	if err := gohassagent.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("agent reset failed: %w", err))
	}
	// Reset registry.
	if err := registry.Reset(gohassagent.GetRegistryPath()); err != nil {
		errs = errors.Join(fmt.Errorf("registry reset failed: %w", err))
	}
	// Reset preferences.
	if err := preferences.Reset(gohassagent.GetPreferencesPath()); err != nil {
		errs = errors.Join(fmt.Errorf("preferences reset failed: %w", err))
	}
	// Reset the log.
	if !ctx.NoLogFile {
		if err := logging.Reset(); err != nil {
			errs = errors.Join(fmt.Errorf("logging reset failed: %w", err))
		}
	}

	if errs != nil {
		slog.Warn("Reset completed with errors", "errors", errs.Error())
	} else {
		slog.Info("Reset completed.")
	}

	return nil
}

type VersionCmd struct{}

func (r *VersionCmd) Run(_ *Context) error {
	fmt.Fprintf(os.Stdout, "%s: %s\n", preferences.AppName, preferences.AppVersion)

	return nil
}

type RegisterCmd struct {
	Server     string `help:"Home Assistant server."`
	Token      string `help:"Personal Access Token."`
	Force      bool   `help:"Force registration."`
	IgnoreURLs bool   `help:"Ignore URLs returned by Home Assistant and use provided server for access."`
}

func (r *RegisterCmd) Help() string {
	return `
Register will attempt to register this device with Home Assistant. Registration
will default to an interactive UI if possible. Details can be provided for
non-interactive registration via the server (--server) and token (--token)
flags. The UI can be explicitly disabled via the --terminal flag.
`
}

func (r *RegisterCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger := logging.New(ctx.LogLevel, ctx.NoLogFile)
	agentCtx = logging.ToContext(agentCtx, logger)

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless),
		agent.WithRegistrationInfo(r.Server, r.Token, r.IgnoreURLs),
		agent.ForceRegister(r.Force))
	if err != nil {
		return fmt.Errorf("failed to run register command: %w", err)
	}

	var trk *sensor.Tracker

	if trk, err = sensor.NewTracker(); err != nil {
		return fmt.Errorf("could not start sensor tracker: %w", err)
	}

	gohassagent.Register(agentCtx, trk)

	return nil
}

type RunCmd struct{}

func (r *RunCmd) Help() string {
	return `
Go Hass Agent reports various device sensors and measurements to, and can
receive desktop notifications from, Home Assistant. It can optionally provide
control of the device via MQTT. It runs as a tray icon application or without
any GUI in a headless mode, processing and sending/receiving data automatically.
The tray icon, if available, provides some actions to configure settings and
show reported sensors/measurements.
`
}

func (r *RunCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger := logging.New(ctx.LogLevel, ctx.NoLogFile)
	agentCtx = logging.ToContext(agentCtx, logger)

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless))
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	var trk *sensor.Tracker

	reg, err := registry.Load(gohassagent.GetRegistryPath())
	if err != nil {
		return fmt.Errorf("could not start registry: %w", err)
	}

	if trk, err = sensor.NewTracker(); err != nil {
		return fmt.Errorf("could not start sensor tracker: %w", err)
	}

	if err := gohassagent.Run(agentCtx, trk, reg); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

//nolint:tagalign
var CLI struct {
	Run       RunCmd       `cmd:"" help:"Run Go Hass Agent."`
	Reset     ResetCmd     `cmd:"" help:"Reset Go Hass Agent."`
	Version   VersionCmd   `cmd:"" help:"Show the Go Hass Agent version."`
	Profile   profileFlags `help:"Enable profiling."`
	AppID     string       `name:"appid" default:"${defaultAppID}" help:"Specify a custom app id (for debugging)."`
	LogLevel  string       `name:"log-level" enum:"info,debug,trace" default:"info" help:"Set logging level."`
	Register  RegisterCmd  `cmd:"" help:"Register with Home Assistant."`
	NoLogFile bool         `help:"Don't write to a log file."`
	Headless  bool         `name:"terminal" help:"Run without a GUI."`
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

	err := ctx.Run(&Context{Headless: CLI.Headless, Profile: CLI.Profile, AppID: CLI.AppID, LogLevel: CLI.LogLevel, NoLogFile: CLI.NoLogFile})
	if CLI.Profile != nil {
		err = errors.Join(logging.StopProfiling(logging.ProfileFlags(CLI.Profile)), err)
	}

	ctx.FatalIfErrorf(err)
}
