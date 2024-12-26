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

type CmdOpts struct {
	Logger   *slog.Logger
	Headless bool
}

type Option func(*CmdOpts)

func AddOptions(options ...Option) *CmdOpts {
	commandOptions := &CmdOpts{}
	for _, option := range options {
		option(commandOptions)
	}

	return commandOptions
}

func RunHeadless(opt bool) Option {
	return func(ctx *CmdOpts) {
		ctx.Headless = opt
	}
}

func WithAppID(id string) Option {
	return func(_ *CmdOpts) {
		preferences.SetAppID(id)
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(ctx *CmdOpts) {
		ctx.Logger = logger
	}
}

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
