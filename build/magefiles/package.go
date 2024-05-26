// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
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
func (Package) Nfpm(arch string) error {
	envMap := GenerateEnv(arch)

	if !FoundOrInstalled("nfpm", "github.com/goreleaser/nfpm/v2/cmd/nfpm@latest") {
		return errors.New("unable to install nfpm")
	}

	for _, pkgformat := range pkgformats {
		slog.Info("Building package with nfpm...", "format", pkgformat)
		args := slices.Concat(nfpmBaseArgs, []string{"--packager", pkgformat})
		if err := sh.RunWithV(envMap, "nfpm", args...); err != nil {
			return err
		}
	}
	return nil
}

// FyneCross builds packages using fyne-cross.
func (Package) FyneCross(arch string) error {
	if !FoundOrInstalled("fyne-cross", "github.com/fyne-io/fyne-cross@latest") {
		return errors.New("unable to install fyne-cross")
	}

	if err := fyneCrossCmd("-arch", arch); err != nil {
		slog.Warn("fyne-cross finished but with errors. Continuing anyway...", "error", err.Error())
	}
	return sh.Copy(
		"fyne-cross/dist/linux-"+arch+"/go-hass-agent-"+arch+".tar.xz",
		"fyne-cross/dist/linux-"+arch+"/go-hass-agent.tar.xz",
	)
}

// CI builds all packages as part of the CI pipeline.
func (p Package) CI(arch string) error {
	if !isCI() {
		return ErrNotCI
	}

	mg.Deps(mg.F(p.Nfpm, arch))
	mg.Deps(mg.F(p.FyneCross, arch))

	return nil
}
