// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"

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

	// Record all licenses in a registry
	mg.Deps(Checks.Licenses)

	return buildProject()
}

// Fast just builds the binary and does not run any prep steps. It will fail if
// the prep steps have not run.
func (Build) Fast() error {
	return buildProject()
}

func (b Build) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.SerialDeps(Preps.Deps)

	mg.SerialDeps(b.Full)

	return nil
}

func buildProject() error {
	envMap, err := GenerateEnv()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	ldflags, err := GetFlags()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	output := "dist/go-hass-agent-" + targetArch

	slog.Info("Running go build...", "output", output, "ldflags", ldflags)

	if err := sh.RunWithV(envMap, "go", "build", "-ldflags="+ldflags, "-o", output); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	return nil
}
