// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package main

import (
	"path/filepath"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type logLevel string

func (d logLevel) AfterApply() error {
	logging.SetLoggingLevel(string(d))
	return nil
}

type profileFlag bool

func (d profileFlag) AfterApply() error {
	logging.SetProfiling()
	return nil
}

type noLogFileFlag bool

func (d noLogFileFlag) AfterApply() error {
	if !d {
		logging.SetLogFile(preferences.LogFile)
	}
	return nil
}

type Context struct {
	AppID    string
	Profile  profileFlag
	Headless bool
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
	a := agent.New(&agent.Options{
		Headless: ctx.Headless,
		ID:       ctx.AppID,
	})
	registry.SetPath(filepath.Join(xdg.ConfigHome, a.AppID(), "sensorRegistry"))
	preferences.SetPath(filepath.Join(xdg.ConfigHome, a.AppID()))
	// Reset agent.
	if err := a.Reset(); err != nil {
		log.Warn().Err(err).Msg("Could not reset agent.")
	}
	// Reset registry.
	if err := registry.Reset(); err != nil {
		log.Warn().Err(err).Msg("Could not reset registry.")
	}
	// Reset preferences.
	if err := preferences.Reset(); err != nil {
		log.Warn().Err(err).Msg("Could not reset preferences.")
	}
	// Reset the log.
	if err := logging.Reset(); err != nil {
		log.Warn().Err(err).Msg("Could not remove log file.")
	}
	log.Info().Msg("Reset complete (refer to any warnings, if any, above.)")
	return nil
}

type VersionCmd struct{}

func (r *VersionCmd) Run(_ *Context) error {
	log.Info().Msgf("%s: %s", preferences.AppName, preferences.AppVersion)
	return nil
}

type RegisterCmd struct {
	Server string `help:"Home Assistant server."`
	Token  string `help:"Personal Access Token."`
	Force  bool   `help:"Force registration."`
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
	a := agent.New(&agent.Options{
		Headless:      ctx.Headless,
		ForceRegister: r.Force,
		Server:        r.Server,
		Token:         r.Token,
		ID:            ctx.AppID,
	})
	var err error

	registry.SetPath(filepath.Join(xdg.ConfigHome, a.AppID(), "sensorRegistry"))
	preferences.SetPath(filepath.Join(xdg.ConfigHome, a.AppID()))
	var trk *sensor.Tracker
	if trk, err = sensor.NewTracker(); err != nil {
		log.Fatal().Err(err).Msg("Could not start sensor sensor.")
	}

	a.Register(trk)
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
	a := agent.New(&agent.Options{
		Headless: ctx.Headless,
		ID:       ctx.AppID,
	})
	var err error

	registry.SetPath(filepath.Join(xdg.ConfigHome, a.AppID(), "sensorRegistry"))
	reg, err := registry.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load sensor registry.")
	}

	preferences.SetPath(filepath.Join(xdg.ConfigHome, a.AppID()))
	var trk *sensor.Tracker
	if trk, err = sensor.NewTracker(); err != nil {
		log.Fatal().Err(err).Msg("Could not start sensor sensor.")
	}

	a.Run(trk, reg)
	return nil
}

var CLI struct {
	Run      RunCmd        `cmd:"" help:"Run Go Hass Agent."`
	Reset    ResetCmd      `cmd:"" help:"Reset Go Hass Agent."`
	Version  VersionCmd    `cmd:"" help:"Show the Go Hass Agent version."`
	AppID    string        `name:"appid" help:"Specify a custom app id (for debugging)."`
	LogLevel logLevel      `name:"log-level" help:"Set logging level."`
	Register RegisterCmd   `cmd:"" help:"Register with Home Assistant."`
	Profile  profileFlag   `help:"Enable profiling."`
	NoLog    noLogFileFlag `help:"Don't write to a log file."`
	Headless bool          `name:"terminal" help:"Run without a GUI."`
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
		log.Fatal().Msg("go-hass-agent should not be run with additional privileges or as root.")
	}
}

func main() {
	kong.Name(preferences.AppName)
	kong.Description(preferences.AppDescription)
	ctx := kong.Parse(&CLI, kong.Bind())
	err := ctx.Run(&Context{Headless: CLI.Headless, Profile: CLI.Profile, AppID: CLI.AppID})
	ctx.FatalIfErrorf(err)
}
