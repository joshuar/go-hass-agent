// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:mnd
package registry

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
)

var gobRegistryFile = "sensor.reg"

type gobRegistry struct {
	sensors map[string]metadata
	mu      sync.Mutex
}

func (g *gobRegistry) write() error {
	regFS, err := os.OpenFile(filepath.Join(registryPath, gobRegistryFile), os.O_RDWR|os.O_CREATE, 0o640)
	if err != nil {
		return fmt.Errorf("could not open registry: %w", err)
	}

	enc := gob.NewEncoder(regFS)

	err = enc.Encode(&g.sensors)
	if err != nil {
		return fmt.Errorf("could not encode registry data: %w", err)
	}

	log.Debug().Msg("Wrote sensor registry to disk.")

	return nil
}

func (g *gobRegistry) read() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	regFS, err := os.OpenFile(filepath.Join(registryPath, gobRegistryFile), os.O_RDWR|os.O_CREATE, 0o640)
	if err != nil {
		return fmt.Errorf("could not open registry: %w", err)
	}

	dec := gob.NewDecoder(regFS)

	err = dec.Decode(&g.sensors)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("could not decode registry data: %w", err)
	}

	log.Debug().Msg("Read sensor registry from disk.")

	return nil
}

func (g *gobRegistry) IsDisabled(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		log.Warn().Str("id", id).Msg("Sensor not found in registry.")

		return false
	}

	return sensor.Disabled
}

func (g *gobRegistry) IsRegistered(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		log.Warn().Str("id", id).Msg("Sensor not found in registry.")

		return false
	}

	return sensor.Registered
}

func (g *gobRegistry) SetDisabled(id string, value bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	m := g.sensors[id]
	m.Disabled = value
	g.sensors[id] = m

	if err := g.write(); err != nil {
		return fmt.Errorf("could not write to registry: %w", err)
	}

	return nil
}

func (g *gobRegistry) SetRegistered(id string, value bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	m := g.sensors[id]
	m.Registered = value
	g.sensors[id] = m

	if err := g.write(); err != nil {
		return fmt.Errorf("could not write to registry: %w", err)
	}

	return nil
}

//revive:disable:unexported-return
func Load() (*gobRegistry, error) {
	reg := &gobRegistry{
		sensors: make(map[string]metadata),
		mu:      sync.Mutex{},
	}
	pathErr := os.MkdirAll(registryPath, 0o755)

	if pathErr != nil && !errors.Is(pathErr, fs.ErrExist) {
		return nil, fmt.Errorf("could not load registry: %w", pathErr)
	}

	if err := reg.read(); err != nil {
		return nil, fmt.Errorf("could not read from registry: %w", err)
	}

	return reg, nil
}
