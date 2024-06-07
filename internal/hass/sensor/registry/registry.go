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

	"github.com/adrg/xdg"
)

var registryPath = filepath.Join(xdg.ConfigHome, "sensorRegistry")

var (
	ErrNotFound        = errors.New("sensor not found")
	ErrInvalidMetadata = errors.New("invalid sensor metadata")
)

func SetPath(path string) {
	registryPath = path
}

func GetPath() string {
	return registryPath
}

func Reset() error {
	err := os.RemoveAll(registryPath)
	if err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	return nil
}
