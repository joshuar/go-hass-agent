//go:build mage
// +build mage

// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/magefile/mage/sh"
)

const (
	pkgBase = "github.com/joshuar/go-hass-agent/internal/preferences"
)

// ------------------------------------------------------------
// Helper Functions
// ------------------------------------------------------------

// FoundOrInstalled checks for existence then installs a file if it's not there
func FoundOrInstalled(executableName, installURL string) (isInstalled bool) {
	_, missing := exec.LookPath(executableName)
	if missing != nil {
		fmt.Printf("installing %v...\n", executableName)
		err := sh.Run("go", "install", installURL)
		if err != nil {
			fmt.Printf("Could not install %v, skipping...\n", executableName)
			return false
		}
		fmt.Printf("%v installed...\n", executableName)
	}
	return true
}

// GetFlags gets all the compile flags to set the version and stuff
func GetFlags() string {
	var flags strings.Builder
	flags.WriteString("-X " + pkgBase + ".gitVersion=" + GitVersion())
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".gitCommit=" + GitHash())
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".buildDate=" + BuildDate())
	return flags.String()
}

func GitVersion() string {
	version, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		slog.Warn("failed to retrieve git version", "error", err.Error())
		return "unknown"
	}
	return version
}

// hash returns the git hash for the current repo or "" if none.
func GitHash() string {
	hash, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		slog.Warn("failed to retrieve git hash", "error", err.Error())
		return "HEAD"
	}
	return hash
}

func BuildDate() string {
	date, err := sh.Output("git", "log", "--date=iso8601-strict", "-1", "--pretty=%ct")
	if err != nil {
		slog.Warn("failed to retrieve build date", "error", err.Error())
		return "unknown"
	}
	return date
}

func buildProject(arch string) error {
	slog.Info("Running go mod download...")
	if err := sh.RunV("go", "mod", "download"); err != nil {
		return err
	}

	slog.Info("Running go build...")
	if err := sh.RunV("go", "build", "-ldflags="+GetFlags(), "-o", "go-hass-agent-"+arch); err != nil {
		return err
	}
	return nil
}
