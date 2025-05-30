// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package cmd contains methods for handling the command-line of the agent.
package cmd

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

const (
	assetsPath = "assets"
	assetsExt  = ".txt"
)

//go:embed assets
var content embed.FS

// API contains shared api objects.
type API struct {
	hass *hass.Client
}

func (a *API) Hass() *hass.Client {
	return a.hass
}

// Opts are the global command-line options common across all commands.
type Opts struct {
	Headless bool
	Path     string
}

// Option represents a command-line option.
type Option func(*Opts) *Opts

// AddOptions adds the given options to a command.
func AddOptions(options ...Option) *Opts {
	commandOptions := &Opts{}
	for _, option := range options {
		option(commandOptions)
	}

	return commandOptions
}

// RunHeadless sets the headless command-line option.
func RunHeadless(opt bool) Option {
	return func(ctx *Opts) *Opts {
		ctx.Headless = opt
		return ctx
	}
}

// WithPath sets a custom path for Go Hass Agent preferences, logs and data.
func WithPath(path string) Option {
	return func(ctx *Opts) *Opts {
		ctx.Path = path
		return ctx
	}
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

func showHelpTxt(file string) string {
	assetFile := filepath.Join(assetsPath, file+assetsExt)

	helpTxt, err := content.ReadFile(assetFile)
	if err != nil {
		slog.Error("Cannot read help text.", slog.Any("error", err))
	}

	return string(helpTxt)
}
