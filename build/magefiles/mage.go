// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
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
	version, _ := getVersion()
	hash, _ := getGitHash()

	platform, arch, ver := parseBuildPlatform()

	// set global logger with custom options
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
	})).
		With(
			slog.Group("git",
				slog.String("version", version),
				slog.String("hash", hash),
			),
			slog.Group("build",
				slog.String("os", platform),
				slog.String("arch", arch),
				slog.String("revision", ver),
			),
		)

	slog.SetDefault(logger)
}
