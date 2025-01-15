// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=state -output state_generated.go -linecomment
const (
	disabledState   state = iota + 1 // disabled
	registeredState                  // registered
)

type state int

var (
	ErrNotFound        = errors.New("sensor not found")
	ErrInvalidMetadata = errors.New("invalid sensor metadata")
)

type metadata struct {
	Registered bool `json:"registered"`
	Disabled   bool `json:"disabled"`
}

// Reset will handle resetting the registry.
func Reset(ctx context.Context) error {
	path := filepath.Join(preferences.PathFromCtx(ctx), "sensorRegistry")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("registry not found: %w", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	return nil
}
