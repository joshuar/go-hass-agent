// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log/slog"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Tests mg.Namespace

// Test runs go test on the project.
func (Tests) Test() error {
	ldflags, err := getFlags()
	if err != nil {
		return fmt.Errorf("cannot run test: %w", err)
	}

	slog.Info("Running go test...")

	if err := sh.RunV("go", "test", "-ldflags="+ldflags, "-coverprofile=coverage.txt", "-v", "./..."); err != nil {
		return fmt.Errorf("failed to run go test: %w", err)
	}

	return nil
}

// Benchmark runs go test -bench on the project.
func (Tests) Benchmark() error {
	slog.Info("Running go test -bench...")

	if err := sh.RunV("go", "test", "-bench=.", "./..."); err != nil {
		return fmt.Errorf("failed to run go benchmarks: %w", err)
	}

	return nil
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
