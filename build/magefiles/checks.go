// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
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
	slog.Info("Running linter (go vet)...")
	if err := sh.RunV("golangci-lint", "run"); err != nil {
		slog.Warn("Linter reported issues.", "error", err.Error())
	}

	if FoundOrInstalled("staticcheck", "honnef.co/go/tools/cmd/staticcheck@latest") {
		slog.Info("Running linter (staticcheck)...")
		if err := sh.RunV("staticcheck", "-f", "stylish", "./..."); err != nil {
			return err
		}
	}

	return nil
}

// Licenses pulls down any dependent project licenses, checking for "forbidden ones".
func (Checks) Licenses() error {
	slog.Info("Running go-licenses...")

	// Make the directory for the license files
	err := os.MkdirAll("licenses", os.ModePerm)
	if err != nil {
		return err
	}

	// If go-licenses is missing, install it
	if FoundOrInstalled("go-licenses", "github.com/google/go-licenses@latest") {
		// The header sets the columns for the contents
		csvHeader := "Package,URL,License\n"

		csvContents, err := sh.Output("go-licenses", "report", "--ignore=github.com/joshuar", "./...")
		if err != nil {
			return err
		}

		// Write out the CSV file with the header row
		if err := os.WriteFile("./licenses/licenses.csv", []byte(csvHeader+csvContents+"\n"), 0o600); err != nil {
			return err
		}
	}

	return nil
}
