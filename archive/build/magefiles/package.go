// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Package mg.Namespace

var (
	pkgformats      = []string{"rpm", "deb", "archlinux"}
	pkgPath         = filepath.Join(distPath, "pkg")
	nfpmCommandLine = []string{
		"go",
		"run",
		"github.com/goreleaser/nfpm/v2/cmd/nfpm",
		"package",
		"--config",
		".nfpm.yaml",
		"--target",
		pkgPath,
	}
)

// Nfpm builds packages using nfpm.
func (Package) Nfpm() error {
	if err := cleanDir(pkgPath); err != nil {
		return fmt.Errorf("cannot run nfpm: %w", err)
	}

	pkgEnv, err := generatePkgEnv()
	if err != nil {
		return fmt.Errorf("cannot run nfpm: %w", err)
	}

	for _, pkgformat := range pkgformats {
		slog.Info("Building package with nfpm.", slog.String("format", pkgformat))
		args := slices.Concat(nfpmCommandLine[1:], []string{"--packager", pkgformat})

		if err := sh.RunWithV(pkgEnv, nfpmCommandLine[0], args...); err != nil {
			return fmt.Errorf("could not run nfpm: %w", err)
		}

		// nfpm creates the same package name for armv6 and armv7 deb packages,
		// so we need to rename them.
		if strings.Contains(pkgEnv["NFPM_ARCH"], "arm") && pkgEnv["NFPM_ARCH"] != "arm64" && pkgformat == "deb" {
			slog.Info("Performing post-packaging steps for arm.")

			debPkgs, err := filepath.Glob(distPath + "/pkg/*.deb")
			if err != nil || debPkgs == nil {
				return fmt.Errorf("could not find arm deb package: %w", err)
			}

			oldDebPkg := debPkgs[0]
			newDebPkg := strings.ReplaceAll(oldDebPkg, "armhf", pkgEnv["NFPM_ARCH"]+"hf")

			if err = sh.Copy(newDebPkg, oldDebPkg); err != nil {
				return fmt.Errorf("could not rename old arm deb package: %w", err)
			}

			err = sh.Rm(oldDebPkg)
			if err != nil {
				return fmt.Errorf("could not remove old arm deb package: %w", err)
			}
		}
	}

	return nil
}

// CI builds all packages as part of the CI pipeline.
func (p Package) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.Deps(p.Nfpm)

	return nil
}
