// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

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

// GobRegistry is a registry based on gob binary data.
type GobRegistry struct {
	sensors map[string]metadata
	file    string
	mu      sync.Mutex
}

func (g *GobRegistry) IsDisabled(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		slog.Debug("Sensor not found in registry.", slog.String("sensor_id", id))

		return false
	}

	return sensor.Disabled
}

func (g *GobRegistry) IsRegistered(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sensor, ok := g.sensors[id]
	if !ok {
		slog.Debug("Sensor not found in registry.", slog.String("sensor_id", id))

		return false
	}

	return sensor.Registered
}

func (g *GobRegistry) SetDisabled(id string, value bool) error {
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

func (g *GobRegistry) SetRegistered(id string, value bool) error {
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

// Load will load the registry from disk.
func Load(path string) (*GobRegistry, error) {
	registryPath := filepath.Join(path, "sensorRegistry", registryFile)

	reg := &GobRegistry{
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

func (g *GobRegistry) write() error {
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

func (g *GobRegistry) read() error {
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
