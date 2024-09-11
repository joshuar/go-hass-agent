// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cli

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	assetsPath = "assets"
	assetsExt  = ".txt"
)

//go:embed assets
var content embed.FS

type Context struct {
	Logger   *slog.Logger
	AppID    string
	Headless bool
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

func WithAppID(id string) CtxOption {
	return func(ctx *Context) {
		ctx.AppID = id
	}
}

func WithLogger(logger *slog.Logger) CtxOption {
	return func(ctx *Context) {
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
