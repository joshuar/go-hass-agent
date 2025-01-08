// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	assetsPath = "assets"
	assetsExt  = ".txt"
)

//go:embed assets
var content embed.FS

// CmdOpts are the global command-line options common across all commands.
type CmdOpts struct {
	Logger   *slog.Logger
	Headless bool
}

// Option represents a command-line option.
type Option func(*CmdOpts) *CmdOpts

// AddOptions adds the given options to a command.
func AddOptions(options ...Option) *CmdOpts {
	commandOptions := &CmdOpts{}
	for _, option := range options {
		option(commandOptions)
	}

	return commandOptions
}

// RunHeadless sets the headless command-line option.
func RunHeadless(opt bool) Option {
	return func(ctx *CmdOpts) *CmdOpts {
		ctx.Headless = opt
		return ctx
	}
}

// WithAppID sets the given ID as the application ID. Useful during development
// and debugging to change the path to the agent preferences.
func WithAppID(id string) Option {
	return func(ctx *CmdOpts) *CmdOpts {
		preferences.SetAppID(id)
		return ctx
	}
}

// WithLogger sets the logger that will be inherited by the command.
func WithLogger(logger *slog.Logger) Option {
	return func(ctx *CmdOpts) *CmdOpts {
		ctx.Logger = logger
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
