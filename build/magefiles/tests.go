// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"log/slog"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Tests mg.Namespace

// Test runs go test on the project.
func (Tests) Test() error {
	slog.Info("Running go test...")
	return sh.RunV("go", "test", "-ldflags="+GetFlags(), "-coverprofile=coverage.txt", "-v", "./...")
}

// Benchmark runs go test -bench on the project.
func (Tests) Benchmark() error {
	slog.Info("Running go test -bench...")
	return sh.RunV("go", "test", "-bench=.", "./...")
}

// CI runs tests as part of a CI pipeline.
func (t Tests) CI(arch string) error {
	if !isCI() {
		return ErrNotCI
	}

	mg.SerialDeps(mg.F(Preps.Deps, arch))

	mg.SerialDeps(t.Test)

	return nil
}
