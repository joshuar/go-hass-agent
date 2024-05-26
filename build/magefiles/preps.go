// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log/slog"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Preps mg.Namespace

var generators = map[string]string{
	"moq":      "github.com/matryer/moq@latest",
	"gotext":   "golang.org/x/text/cmd/gotext@latest",
	"stringer": "golang.org/x/tools/cmd/stringer@latest",
}

// Tidy runs go mod tidy to update the go.mod and go.sum files.
func (Preps) Tidy() error {
	slog.Info("Running go mod tidy...")
	return sh.Run("go", "mod", "tidy")
}

// Format prettifies your code in a standard way to prevent arguments over curly braces.
func (Preps) Format() error {
	slog.Info("Running go fmt...")
	return sh.RunV("go", "fmt", "./...")
}

// Generate ensures all machine-generated files (gotext, stringer, moq, etc.) are up to date.
func (Preps) Generate() error {
	for tool, url := range generators {
		if !FoundOrInstalled(tool, url) {
			return fmt.Errorf("unable to install %s", tool)
		}
	}

	slog.Info("Running go generate...")
	return sh.RunV("go", "generate", "-v", "./...")
}

// BuildDeps installs build dependencies.
func (Preps) Deps(arch string) error {
	if err := sh.RunV("build/scripts/enable-multiarch", arch); err != nil {
		return err
	}
	return sh.RunV("build/scripts/install-deps", arch)
}
