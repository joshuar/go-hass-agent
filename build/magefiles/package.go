// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Package mg.Namespace

const iconPath = "internal/agent/ui/assets/go-hass-agent.png"

var (
	pkgformats   = []string{"rpm", "deb", "archlinux"}
	nfpmBaseArgs = []string{"package", "--config", ".nfpm.yaml", "--target", "dist"}
	fyneCrossCmd = sh.RunCmd("fyne-cross", "linux", "-name", "go-hass-agent", "-icon", iconPath, "-release")

	ErrNoBuildEnv = errors.New("no build and/or version environment variables")
)

// Nfpm builds packages using nfpm.
func (Package) Nfpm() error {
	if !FoundOrInstalled("nfpm", "github.com/goreleaser/nfpm/v2/cmd/nfpm@latest") {
		return errors.New("unable to install nfpm")
	}

	envMap, err := GenerateEnv()
	if err != nil {
		return fmt.Errorf("unable to run nfpm: %w", err)
	}

	for _, pkgformat := range pkgformats {
		slog.Info("Building package with nfpm.", "format", pkgformat)
		args := slices.Concat(nfpmBaseArgs, []string{"--packager", pkgformat})
		if err := sh.RunWithV(envMap, "nfpm", args...); err != nil {
			return err
		}
	}
	return nil
}

// FyneCross builds packages using fyne-cross.
func (Package) FyneCross() error {
	if !FoundOrInstalled("fyne-cross", "github.com/fyne-io/fyne-cross@latest") {
		return errors.New("unable to install fyne-cross")
	}

	envMap, err := GenerateEnv()
	if err != nil {
		return fmt.Errorf("unable to run nfpm: %w", err)
	}

	if err := sh.RunWithV(envMap, "fyne-cross", "linux", "-name", "go-hass-agent", "-icon", iconPath, "-release", "-arch", targetArch); err != nil {
		slog.Warn("fyne-cross finished but with errors. Continuing anyway.", "error", err.Error())
	}
	if err := sh.Copy(
		"fyne-cross/dist/linux-"+targetArch+"/go-hass-agent-"+targetArch+".tar.xz",
		"fyne-cross/dist/linux-"+targetArch+"/go-hass-agent.tar.xz",
	); err != nil {
		return err
	}
	return sh.Rm("fyne-cross/dist/linux-" + targetArch + "/go-hass-agent.tar.xz")
}

// CI builds all packages as part of the CI pipeline.
func (p Package) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.Deps(p.Nfpm)
	mg.Deps(p.FyneCross)

	return nil
}
