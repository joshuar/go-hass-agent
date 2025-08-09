// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

type Preps mg.Namespace

const (
	buildDepsInstallScript = "build/scripts/install-build-deps"
	multiarchScript        = "build/scripts/enable-multiarch"
)

var (
	ErrMissingBuildPlatform = errors.New("BUILDPLATFORM environment variable not set")
	ErrMissingID            = errors.New("ID missing from os-release file")
)

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

// Generate ensures all machine-generated files (gotext, stringer, moq, etc.)
// are up to date.
func (Preps) Generate() error {
	slog.Info("Running go generate...")

	if err := sh.RunV("go", "generate", "./..."); err != nil {
		return fmt.Errorf("failed to run go generate: %w", err)
	}

	return nil
}

// BuildDeps installs build dependencies.
func (Preps) Deps() error {
	buildPlatform, found := os.LookupEnv(platformENV)
	if !found {
		return ErrMissingBuildPlatform
	}

	osrelease, err := whichdistro.GetOSRelease()
	if err != nil {
		return fmt.Errorf("cannot infer distro details: %w", err)
	}

	distroID, found := osrelease.GetValue("ID")
	if !found {
		return ErrMissingID
	}

	if distroID == "ubuntu" {
		if err := sudoWrap(multiarchScript, buildPlatform, distroID); err != nil {
			return fmt.Errorf("unable to enable multiarch for %s: %w", buildPlatform, err)
		}
	}

	if err := sudoWrap(buildDepsInstallScript, buildPlatform, distroID); err != nil {
		return fmt.Errorf("unable to install build deps for %s: %w", buildPlatform, err)
	}

	return nil
}
