// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Checks mg.Namespace

// ------------------------------------------------------------
// Targets for the Magefile that do the quality checks.
// ------------------------------------------------------------

// Lint runs various static checkers to ensure you follow The Rules(tm).
func (Checks) Lint() error {
	slog.Info("Running linter (golangci-lint)...")

	if err := sh.RunV("golangci-lint", "run"); err != nil {
		slog.Warn("Linter reported issues.", "error", err.Error())
	}

	slog.Info("Running linter (staticcheck)...")

	if err := foundOrInstalled("staticcheck", "honnef.co/go/tools/cmd/staticcheck@latest"); err != nil {
		return fmt.Errorf("could not install staticcheck: %w", err)
	}

	if err := sh.RunV("staticcheck", "-f", "stylish", "./..."); err != nil {
		return fmt.Errorf("failed to run staticcheck: %w", err)
	}

	slog.Info("Running linter (nilaway)...")

	if err := foundOrInstalled("nilaway", "go.uber.org/nilaway/cmd/nilaway@latest"); err != nil {
		return fmt.Errorf("could not install nilaway: %w", err)
	}

	if err := sh.RunV("nilaway", "./..."); err != nil {
		return fmt.Errorf("failed to run nilaway: %w", err)
	}

	return nil
}

// Licenses pulls down any dependent project licenses, checking for "forbidden ones".
//
//nolint:mnd
func (Checks) Licenses() error {
	slog.Info("Running go-licenses...")

	// Make the directory for the license files
	err := os.MkdirAll("licenses", os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	// If go-licenses is missing, install it
	if err = foundOrInstalled("go-licenses", "github.com/google/go-licenses@latest"); err != nil {
		return fmt.Errorf("could not install go-licenses: %w", err)
	}
	// The header sets the columns for the contents
	csvHeader := "Package,URL,License\n"

	csvContents, err := sh.Output("go-licenses", "report", "--ignore=github.com/joshuar", "./...")
	if err != nil {
		return fmt.Errorf("could not run go-licenses: %w", err)
	}

	// Write out the CSV file with the header row
	if err := os.WriteFile("./licenses/licenses.csv", []byte(csvHeader+csvContents+"\n"), 0o600); err != nil {
		return fmt.Errorf("could not write licenses database to file: %w", err)
	}

	return nil
}
