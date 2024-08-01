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

//go:generate stringer -type=state -output state_generated.go -linecomment
const (
	disabledState   state = iota + 1 // disabled
	registeredState                  // registered
)

type state int

var registryPath = filepath.Join(xdg.ConfigHome, "sensorRegistry")

var (
	ErrNotFound        = errors.New("sensor not found")
	ErrInvalidMetadata = errors.New("invalid sensor metadata")
)

type metadata struct {
	Registered bool `json:"registered"`
	Disabled   bool `json:"disabled"`
}

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
