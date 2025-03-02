// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

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
		slog.Warn("Linter reported issues.", slog.Any("error", err))
	}

	return nil
}

func (Checks) StaticCheck() error {
	slog.Info("Running linter (staticcheck)...")

	if err := sh.RunV("go", "run", "honnef.co/go/tools/cmd/staticcheck", "-f", "stylish", "./..."); err != nil {
		return fmt.Errorf("failed to run staticcheck: %w", err)
	}

	return nil
}

func (Checks) Nilaway() error {
	slog.Info("Running linter (nilaway)...")

	if err := sh.RunV("go", "run", "go.uber.org/nilaway/cmd/nilaway", "./..."); err != nil {
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

	// The header sets the columns for the contents
	csvHeader := "Package,URL,License\n"

	csvContents, err := sh.Output("go", "run", "github.com/google/go-licenses", "report", "--ignore=github.com/joshuar", "./...")
	if err != nil {
		return fmt.Errorf("could not run go-licenses: %w", err)
	}

	// Write out the CSV file with the header row
	if err := os.WriteFile("./licenses/licenses.csv", []byte(csvHeader+csvContents+"\n"), 0o600); err != nil {
		return fmt.Errorf("could not write licenses database to file: %w", err)
	}

	return nil
}

// All performs all checks.
func (c Checks) All() error {
	mg.Deps(c.Lint(), c.Nilaway(), c.StaticCheck(), c.Licenses())
	return nil
}
