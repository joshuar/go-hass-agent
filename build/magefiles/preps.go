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

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Preps mg.Namespace

const (
	buildDepsInstallScript = "build/scripts/install-build-deps"
	multiarchScript        = "build/scripts/enable-multiarch"
)

var ErrMissingBuildPlatform = errors.New("BUILDPLATFORM environment variable not set")

// Tidy runs go mod tidy to update the go.mod and go.sum files.
func (Preps) Tidy() error {
	slog.Info("Running go mod tidy...")

	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	return nil
}

// Format prettifies your code in a standard way to prevent arguments over curly braces.
func (Preps) Format() error {
	slog.Info("Running go fmt...")

	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		return fmt.Errorf("failed to run go fmt: %w", err)
	}

	return nil
}

// Generate ensures all machine-generated files (gotext, stringer, moq, etc.) are up to date.
func (Preps) Generate() error {
	envMap, err := generateEnv()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	slog.Info("Running go generate...")

	if err := sh.RunWithV(envMap, "go", "generate", "-v", "./..."); err != nil {
		return fmt.Errorf("failed to run go generate: %w", err)
	}

	return nil
}

// BuildDeps installs build dependencies.
func (Preps) Deps() error {
	buildPlatform, found := os.LookupEnv("BUILDPLATFORM")
	if !found {
		return ErrMissingBuildPlatform
	}

	if err := sudoWrap(multiarchScript, buildPlatform); err != nil {
		return fmt.Errorf("unable to enable multiarch for %s: %w", buildPlatform, err)
	}

	if err := sudoWrap(buildDepsInstallScript, buildPlatform); err != nil {
		return fmt.Errorf("unable to install build deps for %s: %w", buildPlatform, err)
	}

	return nil
}
