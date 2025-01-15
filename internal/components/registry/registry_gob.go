// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	registryFile     = "sensor.reg"
	defaultFilePerms = 0o640
)

type gobRegistry struct {
	sensors map[string]metadata
	file    string
	mu      sync.Mutex
}

func (g *gobRegistry) write() error {
	regFS, err := os.OpenFile(g.file, os.O_RDWR|os.O_CREATE, defaultFilePerms)
	if err != nil {
		return fmt.Errorf("could not open registry for writing: %w", err)
	}

	enc := gob.NewEncoder(regFS)

	err = enc.Encode(&g.sensors)
	if err != nil {
		return fmt.Errorf("could not encode registry data: %w", err)
	}

	return nil
}

func (g *gobRegistry) read() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	regFS, err := os.OpenFile(g.file, os.O_RDWR|os.O_CREATE, defaultFilePerms)
	if err != nil {
		return fmt.Errorf("could not open registry for reading: %w", err)
	}

	dec := gob.NewDecoder(regFS)

	err = dec.Decode(&g.sensors)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("could not decode registry data: %w", err)
	}

	return nil
}

func (g *gobRegistry) IsDisabled(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		slog.Warn("Sensor not found in registry.", slog.String("sensor_id", id))

		return false
	}

	return sensor.Disabled
}

func (g *gobRegistry) IsRegistered(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		slog.Warn("Sensor not found in registry.", slog.String("sensor_id", id))

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
func Load(path string) (*gobRegistry, error) {
	registryPath := filepath.Join(path, "sensorRegistry", registryFile)

	reg := &gobRegistry{
		sensors: make(map[string]metadata),
		mu:      sync.Mutex{},
		file:    registryPath,
	}

	if err := checkPath(filepath.Dir(reg.file)); err != nil {
		return nil, fmt.Errorf("could not load registry: %w", err)
	}

	if err := reg.read(); err != nil {
		return nil, fmt.Errorf("could not read from registry: %w", err)
	}

	return reg, nil
}
