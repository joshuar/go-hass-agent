// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package registry handles managing a sensor registry locally on disk.
package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrNotFound is returned when a sensor could not be found in the registry.
	ErrNotFound = errors.New("sensor not found")
	// ErrInvalidMetadata is returned when the sensor data in the registry is invalid.
	ErrInvalidMetadata = errors.New("invalid sensor metadata")
)

type metadata struct {
	Registered bool `json:"registered"`
	Disabled   bool `json:"disabled"`
}

// Reset will handle resetting the registry.
func Reset(registryPath string) error {
	registryPath = filepath.Join(registryPath, "sensorRegistry")

	_, err := os.Stat(registryPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("registry not found: %w", err)
	}

	err = os.RemoveAll(registryPath)
	if err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	return nil
}
