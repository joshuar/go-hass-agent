// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=state -output metadataStates.go -linecomment
const (
	disabledState   state = iota + 1 // disabled
	registeredState                  // registered
)

type state int

type jsonFilesRegistry struct {
	sensors sync.Map
	path    string
}

type metadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}

func (j *jsonFilesRegistry) get(id string, valueType state) bool {
	var meta metadata
	var value any
	var ok bool
	if value, ok = j.sensors.Load(id); !ok {
		log.Warn().Str("sensor", id).Str("metadata", valueType.String()).
			Msg("Sensor metadata not found.")
		return false
	}
	if meta, ok = value.(metadata); !ok {
		log.Warn().Str("sensor", id).Msg("Invalid sensor metadata.")
		return false
	}
	switch valueType {
	case disabledState:
		return meta.Disabled
	case registeredState:
		return meta.Registered
	}
	return false
}

func (j *jsonFilesRegistry) set(id string, valueType state, value bool) error {
	var m metadata
	if v, ok := j.sensors.Load(id); !ok {
		log.Warn().Str("sensor", id).Msg("Sensor not found in registry. Will add as new.")
	} else {
		var ok bool
		if m, ok = v.(metadata); !ok {
			log.Warn().Str("sensor", id).Str("metadata", valueType.String()).
				Msg("Sensor metadata invalid. Ignoring.")
		}
	}
	switch valueType {
	case disabledState:
		m.Disabled = value
	case registeredState:
		m.Registered = value
	}
	j.sensors.Store(id, m)
	err := j.write(id)
	if err != nil {
		return err
	}
	return nil
}

func (j *jsonFilesRegistry) write(id string) error {
	var v any
	var m metadata
	var ok bool
	path := j.path + "/" + id + ".json"
	if v, ok = j.sensors.Load(id); !ok {
		return errors.New("not found")
	}
	if m, ok = v.(metadata); !ok {
		return errors.New("invalid metadata")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func (j *jsonFilesRegistry) IsDisabled(id string) bool {
	return j.get(id, disabledState)
}

func (j *jsonFilesRegistry) IsRegistered(id string) bool {
	return j.get(id, registeredState)
}

func (j *jsonFilesRegistry) SetDisabled(id string, value bool) error {
	return j.set(id, disabledState, value)
}

func (j *jsonFilesRegistry) SetRegistered(id string, value bool) error {
	return j.set(id, registeredState, value)
}

func Load() (*jsonFilesRegistry, error) {
	reg := &jsonFilesRegistry{
		sensors: sync.Map{},
		path:    registryPath,
	}
	pathErr := os.Mkdir(registryPath, 0o755)
	if pathErr != nil && !errors.Is(pathErr, fs.ErrExist) {
		return nil, pathErr
	}
	files, err := filepath.Glob(registryPath + "/*.json")
	if err != nil {
		return nil, err
	}
	go func() {
		for _, filename := range files {
			id, meta := parseFile(filename)
			reg.sensors.Store(id, meta)
		}
	}()
	return reg, nil
}

func parseFile(path string) (string, metadata) {
	sensorID, _ := strings.CutSuffix(filepath.Base(path), ".json")
	log.Trace().Msgf("Getting information from registry for %s", sensorID)
	b, err := os.ReadFile(path)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to read contents of %s. Skipping", path)
		return "", metadata{}
	}
	m := &metadata{}
	err = json.Unmarshal(b, m)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to parse %s. Skipping.", path)
		return "", metadata{}
	}
	return sensorID, *m
}
