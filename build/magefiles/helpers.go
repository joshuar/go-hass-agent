// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/magefile/mage/sh"
)

const (
	pkgBase = "github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrNotCI = errors.New("not in CI environment")

// isCI checks whether we are currently running as part of a CI pipeline (i.e.
// in a GitHub runner).
func isCI() bool {
	return os.Getenv("CI") != ""
}

// isRoot checks whether we are running as the root user or with elevated
// privileges.
func isRoot() bool {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	gid := syscall.Getgid()
	if uid != euid || gid != egid || uid == 0 {
		return true
	}
	return false
}

// FoundOrInstalled checks for existence then installs a file if it's not there.
func FoundOrInstalled(executableName, installURL string) (isInstalled bool) {
	_, missing := exec.LookPath(executableName)
	if missing != nil {
		slog.Info("Installing build tool...", "tool", executableName, "url", installURL)
		err := sh.Run("go", "install", installURL)
		if err != nil {
			slog.Warn("Could not install tool, skipping...", "tool", executableName)
			return false
		}
		slog.Info("Tool installed...", "tool", executableName)
	}
	return true
}

// GetFlags gets all the compile flags to set the version and stuff.
func GetFlags() string {
	var flags strings.Builder
	flags.WriteString("-X " + pkgBase + ".gitVersion=" + GitVersion())
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".gitCommit=" + GitHash())
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".buildDate=" + BuildDate())
	return flags.String()
}

// GitVersion returns a string that can be used as a version string.
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

// BuildDate returns the build date.
func BuildDate() string {
	date, err := sh.Output("git", "log", "--date=iso8601-strict", "-1", "--pretty=%ct")
	if err != nil {
		slog.Warn("failed to retrieve build date", "error", err.Error())
		return "unknown"
	}
	return date
}

// GenerateEnv will create a map[string]string containing environment variables
// and their values necessary for building the package on the given
// architecture.
func GenerateEnv(arch string) map[string]string {
	envMap := make(map[string]string)

	envMap["NFPM_ARCH"] = arch
	envMap["CGO_ENABLED"] = "1"

	version := GitVersion()
	if version == "unknown" || version == "" {
		slog.Warn("Could not retrieve version.")
	}
	envMap["APPVERSION"] = version

	switch arch {
	case "arm":
		envMap["CC"] = "arm-linux-gnueabihf-gcc"
		envMap["PKG_CONFIG_PATH"] = "/usr/lib/arm-linux-gnueabihf/pkgconfig"
		envMap["GOARCH"] = "arm"
		envMap["GOARM"] = "7"
	case "arm64":
		envMap["CC"] = "aarch64-linux-gnu-gcc"
		envMap["PKG_CONFIG_PATH"] = "/usr/lib/aarch64-linux-gnu/pkgconfig"
		envMap["GOARCH"] = arch
	}
	return envMap
}
