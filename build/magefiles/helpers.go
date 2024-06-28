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
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"syscall"

	"github.com/magefile/mage/sh"
)

const (
	pkgBase  = "github.com/joshuar/go-hass-agent/internal/preferences"
	distPath = "dist"
)

var (
	ErrNotCI           = errors.New("not in CI environment")
	ErrUnsupportedArch = errors.New("unsupported target architecture")
)

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

// sudoWrap will "wrap" the given command with sudo if needed.
func sudoWrap(cmd string, args ...string) error {
	if isRoot() {
		if err := sh.RunV(cmd, args...); err != nil {
			return fmt.Errorf("could not run command: %w", err)
		}
	} else {
		if err := sh.RunV("sudo", slices.Concat([]string{cmd}, args)...); err != nil {
			return fmt.Errorf("could not run command: %w", err)
		}
	}

	return nil
}

// foundOrInstalled checks for existence then installs a file if it's not there.
func foundOrInstalled(executableName, installURL string) error {
	_, missing := exec.LookPath(executableName)
	if missing != nil {
		slog.Info("Installing tool.", "tool", executableName, "url", installURL)

		err := sh.Run("go", "install", installURL)
		if err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
	}

	return nil
}

// getFlags gets all the compile flags to set the version and stuff.
func getFlags() (string, error) {
	var version, hash, date string

	var err error

	if version, err = getVersion(); err != nil {
		return "", fmt.Errorf("failed to retrieve version from git: %w", err)
	}

	if hash, err = getGitHash(); err != nil {
		return "", fmt.Errorf("failed to retrieve hash from git: %w", err)
	}

	if date, err = getBuildDate(); err != nil {
		return "", fmt.Errorf("failed to retrieve build date from git: %w", err)
	}

	var flags strings.Builder

	flags.WriteString("-X " + pkgBase + ".gitVersion=" + version)
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".gitCommit=" + hash)
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgBase + ".buildDate=" + date)

	return flags.String(), nil
}

// getVersion returns a string that can be used as a version string.
func getVersion() (string, error) {
	// Use the version already set in the environment (i.e., by the CI run).
	if version, ok := os.LookupEnv("APPVERSION"); ok {
		return version, nil
	}
	// Else, derive a version from git.
	version, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return "", fmt.Errorf("could not get version from git: %w", err)
	}

	return version, nil
}

// hash returns the git hash for the current repo or "" if none.
func getGitHash() (string, error) {
	hash, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not get git hash: %w", err)
	}

	return hash, nil
}

// getBuildDate returns the build date.
func getBuildDate() (string, error) {
	date, err := sh.Output("git", "log", "--date=iso8601-strict", "-1", "--pretty=%ct")
	if err != nil {
		return "", fmt.Errorf("could not get build date from git: %w", err)
	}

	return date, nil
}

// generateEnv will create a map[string]string containing environment variables
// and their values necessary for building the package on the given
// architecture.
func generateEnv() (map[string]string, error) {
	envMap := make(map[string]string)

	// CGO_ENABLED is required.
	envMap["CGO_ENABLED"] = "1"

	version, err := getVersion()
	if err != nil {
		return nil, fmt.Errorf("could not generate environment: %w", err)
	}
	// Set APPVERSION to current version.
	envMap["APPVERSION"] = version

	// Set NFPM_ARCH so that nfpm knows how to package for this arch.
	envMap["NFPM_ARCH"] = runtime.GOARCH

	// Get the value of BUILDPLATFORM (if set) from the environment, which
	// indicates cross-compilation has been requested.
	_, arch, ver := parseBuildPlatform()

	if arch != "" && arch != runtime.GOARCH {
		slog.Info("Cross compilation requested.", "host", runtime.GOARCH, "target", arch)

		// Update NFPM_ARCH to the target arch.
		envMap["NFPM_ARCH"] = arch + ver

		// Set additional build-related variables based on the target arch.
		switch arch {
		case "arm":
			envMap["CC"] = "arm-linux-gnueabihf-gcc"
			envMap["PKG_CONFIG_PATH"] = "/usr/lib/arm-linux-gnueabihf/pkgconfig"
			envMap["GOARCH"] = arch
			envMap["GOARM"] = ver
			envMap["PLATFORMPAIR"] = arch + ver
		case "arm64":
			envMap["CC"] = "aarch64-linux-gnu-gcc"
			envMap["PKG_CONFIG_PATH"] = "/usr/lib/aarch64-linux-gnu/pkgconfig"
			envMap["GOARCH"] = arch
			envMap["PLATFORMPAIR"] = arch
		default:
			return nil, ErrUnsupportedArch
		}
	} else {
		envMap["GOARCH"] = runtime.GOARCH
		envMap["PLATFORMPAIR"] = runtime.GOARCH
	}

	return envMap, nil
}

//nolint:mnd
func parseBuildPlatform() (operatingsystem, architecture, version string) {
	var buildPlatform string

	var ok bool

	if buildPlatform, ok = os.LookupEnv("BUILDPLATFORM"); !ok {
		return runtime.GOOS, runtime.GOARCH, ""
	}

	buildComponents := strings.Split(buildPlatform, "/")
	operatingsystem = buildComponents[0]
	architecture = buildComponents[1]

	if len(buildComponents) > 2 {
		version = strings.TrimPrefix(buildComponents[2], "v")
	}

	return operatingsystem, architecture, version
}
