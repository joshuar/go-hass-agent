// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"log/slog"
	"os"
)

const (
	distPath    = "dist"
	platformENV = "TARGETPLATFORM"
)

var (
	ErrNotCI           = errors.New("not in CI environment")
	ErrUnsupportedArch = errors.New("unsupported target architecture")
)

func init() {
	// set global logger with custom options
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}
