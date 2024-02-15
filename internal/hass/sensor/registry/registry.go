// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
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

func Reset() {
	var err error
	if err = os.RemoveAll(registryPath); err != nil {
		log.Warn().Err(err).Msg("Could not remove existing registry.")
		return
	}
	log.Info().Msg("Registry reset.")
}
