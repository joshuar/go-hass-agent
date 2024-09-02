// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cli

import (
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	assetsPath = "assets"
	assetsExt  = ".txt"
)

//go:embed assets
var content embed.FS

type Context struct {
	Profile   ProfileFlags
	AppID     string
	LogLevel  string
	Headless  bool
	NoLogFile bool
}

type CtxOption func(*Context)

func CreateCtx(options ...CtxOption) *Context {
	ctx := &Context{}
	for _, option := range options {
		option(ctx)
	}

	return ctx
}

func RunHeadless(opt bool) CtxOption {
	return func(ctx *Context) {
		ctx.Headless = opt
	}
}

func WithProfileFlags(flags ProfileFlags) CtxOption {
	return func(ctx *Context) {
		ctx.Profile = flags
	}
}

func WithAppID(id string) CtxOption {
	return func(ctx *Context) {
		ctx.AppID = id
	}
}

func WithLogLevel(level string) CtxOption {
	return func(ctx *Context) {
		ctx.LogLevel = level
	}
}

func WithLogFile(opt bool) CtxOption {
	return func(ctx *Context) {
		ctx.NoLogFile = opt
	}
}

type ProfileFlags logging.ProfileFlags

func (d ProfileFlags) AfterApply() error {
	err := logging.StartProfiling(logging.ProfileFlags(d))
	if err != nil {
		return fmt.Errorf("could not start profiling: %w", err)
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
