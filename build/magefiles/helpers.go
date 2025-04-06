// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"

	"github.com/magefile/mage/sh"
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

// getFlags gets all the compile flags to set the version and stuff.
func getFlags() (string, error) {
	// pkgPath is where flags are derived from.
	pkgPath := "github.com/joshuar/go-hass-agent/internal/components/preferences"

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

	flags.WriteString("-X " + pkgPath + ".gitVersion=" + version)
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgPath + ".gitCommit=" + hash)
	flags.WriteString(" ")
	flags.WriteString("-X " + pkgPath + ".buildDate=" + date)

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

// getGitHash returns the git hash for the current repo or "" if none.
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

// generateBuildEnv will create a map[string]string containing environment
// variables and their values necessary for building Go Hass Agent on the given
// architecture.
func generateBuildEnv() (map[string]string, error) {
	envMap := make(map[string]string)

	// CGO_ENABLED is required.
	envMap["CGO_ENABLED"] = "1"

	version, err := getVersion()
	if err != nil {
		return nil, fmt.Errorf("could not generate environment: %w", err)
	}
	// Set APPVERSION to current version.
	envMap["APPVERSION"] = version

	// Get the value of BUILDPLATFORM (if set) from the environment, which
	// indicates cross-compilation has been requested.
	_, arch, ver := parseBuildPlatform()

	if arch != "" && arch != runtime.GOARCH {
		slog.Info("Setting up cross-compilation.")
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

	// Set an appropriate output file based on the arch to build for.
	envMap["OUTPUT"] = filepath.Join(distPath, "/go-hass-agent-"+envMap["PLATFORMPAIR"])

	return envMap, nil
}

// generatePkgEnv will create a map[string]string containing environment
// variables and their values necessary for packaging Go Hass Agent on the given
// architecture.
func generatePkgEnv() (map[string]string, error) {
	envMap := make(map[string]string)

	version, err := getVersion()
	if err != nil {
		return nil, fmt.Errorf("could not generate env: %w", err)
	}
	// Set APPVERSION to current version.
	envMap["APPVERSION"] = version
	// Set NFPM_ARCH so that nfpm knows how to package for this arch.
	envMap["NFPM_ARCH"] = runtime.GOARCH
	// Parse the build platform from the environment.
	_, arch, ver := parseBuildPlatform()
	// For arm, set the NFPM_ARCH to include the revision.
	if arch != "" && arch != runtime.GOARCH {
		slog.Info("Setting up cross-compilation.")
		// Update NFPM_ARCH to the target arch.
		envMap["NFPM_ARCH"] = arch + ver
	}

	return envMap, nil
}

// parseBuildPlatform reads the TARGETPLATFORM environment variable, which should
// always be set, and extracts the value into appropriate GOOS, GOARCH and GOARM
// (if applicable) variables.
func parseBuildPlatform() (string, string, string) {
	var (
		buildPlatform string
		opsys         string
		arch          string
		version       string
		ok            bool
	)

	if buildPlatform, ok = os.LookupEnv(platformENV); !ok {
		return runtime.GOOS, runtime.GOARCH, ""
	}

	buildComponents := strings.Split(buildPlatform, "/")
	opsys = buildComponents[0]

	if len(buildComponents) > 1 {
		arch = buildComponents[1]
	}

	if len(buildComponents) > 2 {
		version = strings.TrimPrefix(buildComponents[2], "v")
	}

	return opsys, arch, version
}

func cleanDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("could not clean directory %s: %w", path, err)
	}

	if err := os.MkdirAll(path, os.ModeAppend); err != nil {
		return fmt.Errorf("could not create directory %s: %w", path, err)
	}

	return nil
}
