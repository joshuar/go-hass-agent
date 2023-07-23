// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
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
	switch valueType {
	case disabledState:
		if value, ok := j.sensors.Load(id); ok {
			return value.(metadata).Disabled
		}
	case registeredState:
		if value, ok := j.sensors.Load(id); ok {
			return value.(metadata).Registered
		}
	}
	return false
}

func (j *jsonFilesRegistry) IsDisabled(id string) chan bool {
	valueCh := make(chan bool, 1)
	defer close(valueCh)
	valueCh <- j.get(id, disabledState)
	return valueCh
}

func (j *jsonFilesRegistry) IsRegistered(id string) chan bool {
	valueCh := make(chan bool, 1)
	defer close(valueCh)
	valueCh <- j.get(id, registeredState)
	return valueCh
}

func (j *jsonFilesRegistry) set(id string, valueType state, value bool) error {
	var m metadata
	if v, ok := j.sensors.Load(id); !ok {
		log.Warn().Msgf("Sensor %s not found in registry. Will add value as new.", id)
	} else {
		m = v.(metadata)
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
	path := j.path + "/" + id + ".json"
	if v, ok := j.sensors.Load(id); ok {
		m := v.(metadata)
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return os.WriteFile(path, b, 0644)
	}
	return errors.New("not found")
}

func (j *jsonFilesRegistry) SetDisabled(id string, value bool) error {
	return j.set(id, disabledState, value)
}

func (j *jsonFilesRegistry) SetRegistered(id string, value bool) error {
	return j.set(id, registeredState, value)
}

func NewJsonFilesRegistry(path string) (*jsonFilesRegistry, error) {
	reg := &jsonFilesRegistry{
		sensors: sync.Map{},
		path:    path,
	}
	pathErr := os.Mkdir(path, 0755)
	if pathErr != nil && !errors.Is(pathErr, fs.ErrExist) {
		return nil, pathErr
	}
	files, err := filepath.Glob(path + "/*.json")
	if err != nil {
		return nil, err
	}
	go func() {
		for _, filename := range files {
			sensorID, metadata := parseFile(filename)
			reg.sensors.Store(sensorID, metadata)
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
