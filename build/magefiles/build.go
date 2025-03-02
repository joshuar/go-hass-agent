// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

var ErrBuildFailed = errors.New("build failed")

// Full runs all prep steps and then builds the binary.
func (Build) Full() error {
	slog.Info("Starting full build.")

	return buildProject()
}

// CI runs build steps required for building in a CI environment (i.e., GitHub).
func (b Build) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.SerialDeps(Preps.Deps)

	mg.SerialDeps(b.Full)

	return nil
}

// buildProject is the shared method that all exported build targets use. It
// runs the bare minimum steps to build a binary of the agent.
func buildProject() error {
	if err := cleanDir(distPath); err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	// Set-up the build environment.
	buildEnv, err := generateBuildEnv()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	// Set-up appropriate build flags.
	ldflags, err := getFlags()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	//nolint:sloglint
	slog.Info("Running go build...",
		slog.String("output", buildEnv["OUTPUT"]),
		slog.String("build.host", runtime.GOARCH))

	// Run the build.
	if err := sh.RunWithV(buildEnv, "go", "build", "-ldflags="+ldflags, "-o", buildEnv["OUTPUT"]); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	return nil
}
