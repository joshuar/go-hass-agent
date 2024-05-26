// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"log/slog"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

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

	mg.SerialDeps(mg.F(Preps.Deps, arch))

	mg.SerialDeps(mg.F(b.Full, arch))
	return nil
}

func buildProject(arch string) error {
	envMap := GenerateEnv(arch)

	slog.Info("Running go build...")
	return sh.RunWithV(envMap, "go", "build", "-ldflags="+GetFlags(), "-o", "dist/go-hass-agent-"+arch)
}
