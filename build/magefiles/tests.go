// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Tests mg.Namespace

// Test runs go test on the project.
func (Tests) Test() error {
	ldflags, err := GetFlags()
	if err != nil {
		return fmt.Errorf("cannot run test: %w", err)
	}

	fmt.Println("Running go test...")
	return sh.RunV("go", "test", "-ldflags="+ldflags, "-coverprofile=coverage.txt", "-v", "./...")
}

// Benchmark runs go test -bench on the project.
func (Tests) Benchmark() error {
	fmt.Println("Running go test -bench...")
	return sh.RunV("go", "test", "-bench=.", "./...")
}

// CI runs tests as part of a CI pipeline.
func (t Tests) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.SerialDeps(Preps.Deps)

	mg.SerialDeps(t.Test)

	return nil
}
