// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"log/slog"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

var ErrBuildFailed = errors.New("build failed")

// Full runs all prep steps and then builds the binary.
func (Build) Full(arch string) error {
	slog.Info("Starting full build", "arch", arch)

	// Make everything nice, neat, and proper
	mg.Deps(Preps.Tidy)
	mg.Deps(Preps.Format)
	mg.Deps(Preps.Generate)

	// Record all licenses in a registry
	mg.Deps(Checks.Licenses)

	return buildProject(arch)
}

// Fast just builds the binary and does not run any prep steps. It will fail if
// the prep steps have not run.
func (Build) Fast(arch string) error {
	return buildProject(arch)
}

func (b Build) CI(arch string) error {
	if !isCI() {
		return ErrNotCI
	}
	arch, err := validateArch(arch)
	if err != nil {
		return err
	}

	mg.SerialDeps(mg.F(Preps.Deps, arch))

	mg.SerialDeps(mg.F(b.Full, arch))
	return nil
}

func buildProject(arch string) error {
	envMap, err := GenerateEnv(arch)
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	ldflags, err := GetFlags()
	if err != nil {
		return errors.Join(ErrBuildFailed, err)
	}

	slog.Info("Running go build...")
	if err := sh.RunWithV(envMap, "go", "build", "-ldflags="+ldflags); err != nil {
		return errors.Join(ErrBuildFailed, err)
	}
	return sh.Copy("dist/go-hass-agent-"+arch, "go-hass-agent")
}
