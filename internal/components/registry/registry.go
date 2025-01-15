// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrNotFound        = errors.New("sensor not found")
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
