// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"os"
	"runtime"

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
	fmt.Println("Running go mod tidy...")
	return sh.Run("go", "mod", "tidy")
}

// Format prettifies your code in a standard way to prevent arguments over curly braces.
func (Preps) Format() error {
	fmt.Println("Running go fmt...")
	return sh.RunV("go", "fmt", "./...")
}

// Generate ensures all machine-generated files (gotext, stringer, moq, etc.) are up to date.
func (Preps) Generate() error {
	for tool, url := range generators {
		if !FoundOrInstalled(tool, url) {
			return fmt.Errorf("unable to install %s", tool)
		}
	}

	fmt.Println("Running go generate...")
	return sh.RunV("go", "generate", "-v", "./...")
}

// BuildDeps installs build dependencies.
func (Preps) Deps() error {
	if v, ok := os.LookupEnv("TARGETARCH"); ok {
		targetArch = v
	}
	if targetArch != "" && targetArch != runtime.GOARCH {
		if err := SudoWrap("build/scripts/enable-multiarch", targetArch); err != nil {
			return fmt.Errorf("unable to enable multiarch for %s: %w", targetArch, err)
		}
		if err := SudoWrap("build/scripts/install-deps", targetArch, runtime.GOARCH); err != nil {
			return fmt.Errorf("unable to enable multiarch for %s: %w", targetArch, err)
		}
	} else {
		if err := SudoWrap("build/scripts/install-deps", runtime.GOARCH); err != nil {
			return fmt.Errorf("unable to enable multiarch for %s: %w", targetArch, err)
		}
	}
	return nil
}
