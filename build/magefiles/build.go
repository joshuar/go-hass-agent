// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

var ErrBuildFailed = errors.New("build failed")

// Full runs all prep steps and then builds the binary.
func (Build) Full() error {
	slog.Info("Starting full build.")

	// Make everything nice, neat, and proper
	mg.Deps(Preps.Tidy)
	mg.Deps(Preps.Format)
	mg.Deps(Preps.Generate)

	return buildProject()
}

// Fast just builds the binary and does not run any prep steps. It will fail if
// the prep steps have not run.
func (Build) Fast() error {
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
//
//nolint:mnd
func buildProject() error {
	// Remove any existing dist directory.
	if err := os.RemoveAll(distPath); err != nil {
		return fmt.Errorf("could not clean dist directory: %w", err)
	}
	// Recreate an empty dist directory for this build.
	if err := os.Mkdir(distPath, 0o755); err != nil {
		return fmt.Errorf("could not create dist directory: %w", err)
	}

	// Set-up the build environment.
	envMap, err := generateEnv()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	// Set-up appropriate build flags.
	ldflags, err := getFlags()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	// Set an appropriate output file based on the arch to build for.
	outputFile := filepath.Join(distPath, "/go-hass-agent-"+envMap["PLATFORMPAIR"])

	slog.Info("Running go build...",
		slog.String("output", outputFile),
		slog.String("ldflags", ldflags))

	// Run the build.
	if err := sh.RunWithV(envMap, "go", "build", "-ldflags="+ldflags, "-o", outputFile); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	return nil
}
