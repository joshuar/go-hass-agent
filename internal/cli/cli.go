// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cli

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joshuar/go-hass-agent/internal/logging"
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
	AppID    string
	Headless bool
}

type Option func(*CmdOpts)

func CreateCtx(options ...Option) *CmdOpts {
	ctx := &CmdOpts{}
	for _, option := range options {
		option(ctx)
	}

	return ctx
}

func RunHeadless(opt bool) Option {
	return func(ctx *CmdOpts) {
		ctx.Headless = opt
	}
}

func WithAppID(id string) Option {
	return func(ctx *CmdOpts) {
		ctx.AppID = id
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

func newContext(opts *CmdOpts) (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = preferences.AppIDToContext(ctx, opts.AppID)

	return ctx, cancelFunc
}

func showHelpTxt(file string) string {
	assetFile := filepath.Join(assetsPath, file+assetsExt)

	helpTxt, err := content.ReadFile(assetFile)
	if err != nil {
		slog.Error("Cannot read help text.", slog.Any("error", err))
	}

	return string(helpTxt)
}
