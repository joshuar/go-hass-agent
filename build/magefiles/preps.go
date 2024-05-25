//go:build mage
// +build mage

// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Preps mg.Namespace

// Tidy runs go mod tidy to update the go.mod and go.sum files
func (Preps) Tidy() error {
	fmt.Println("Running go mod tidy...")
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	return nil
}

// Format prettifies your code in a standard way to prevent arguments over curly braces
func (Preps) Format() error {
	fmt.Println("Running go fmt...")
	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		return err
	}
	return nil
}

// Generate ensures all machine-generated files (gotext, stringer, moq, etc.) are up to date
func (Preps) Generate() error {
	if !FoundOrInstalled("moq", "github.com/matryer/moq@latest") {
		return errors.New("unable to install moq")
	}
	if !FoundOrInstalled("gotext", "golang.org/x/text/cmd/gotext@latest") {
		return errors.New("unable to install gotext")
	}
	if !FoundOrInstalled("stringer", "golang.org/x/tools/cmd/stringer@latest") {
		return errors.New("unable to install stringer")
	}

	fmt.Println("Running go generate...")
	if err := sh.RunV("go", "generate", "-v", "./..."); err != nil {
		return err
	}
	return nil
}
