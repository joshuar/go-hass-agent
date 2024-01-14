// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	registry "github.com/joshuar/go-hass-agent/internal/tracker/registry/jsonFiles"
	"github.com/rs/zerolog/log"
)

var basePath = filepath.Join(os.Getenv("HOME"), ".config")

//go:generate moq -out mock_Registry_test.go . Registry
type Registry interface {
	SetDisabled(string, bool) error
	SetRegistered(string, bool) error
	IsDisabled(string) chan bool
	IsRegistered(string) chan bool
	Path() string
}

//go:generate moq -out mock_apiResponse_test.go . apiResponse
type apiResponse interface {
	Registered() bool
	Disabled() bool
	Type() api.ResponseType
}

type SensorTracker struct {
	registry Registry
	sensor   map[string]Sensor
	mu       sync.Mutex
}

// Add creates a new sensor in the tracker based on a received state update.
func (t *SensorTracker) add(s Sensor) error {
	t.mu.Lock()
	if t.sensor == nil {
		t.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	t.sensor[s.ID()] = s
	t.mu.Unlock()
	return nil
}

// Get fetches a sensors current tracked state.
func (t *SensorTracker) Get(id string) (Sensor, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sensor[id] != nil {
		return t.sensor[id], nil
	} else {
		return nil, errors.New("not found")
	}
}

func (t *SensorTracker) SensorList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sensor == nil {
		log.Warn().Msg("No sensors available.")
		return nil
	}
	sortedEntities := make([]string, 0, len(t.sensor))
	for name, sensor := range t.sensor {
		if sensor.State() != nil {
			sortedEntities = append(sortedEntities, name)
		}
	}
	sort.Strings(sortedEntities)
	return sortedEntities
}

// send will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (t *SensorTracker) send(ctx context.Context, sensorUpdate Sensor) {
	var req api.Request
	if disabled := <-t.registry.IsDisabled(sensorUpdate.ID()); disabled {
		log.Debug().Str("id", sensorUpdate.ID()).
			Msg("Sensor is disabled. Ignoring update.")
		return
	}
	registered := <-t.registry.IsRegistered(sensorUpdate.ID())
	req = marshallSensorState(sensorUpdate, registered)
	response := <-api.ExecuteRequest(ctx, req)
	switch r := response.(type) {
	case apiResponse:
		t.handle(r, sensorUpdate)
	case error:
		log.Warn().Err(r).Str("id", sensorUpdate.ID()).
			Msg("Failed to send sensor data to Home Assistant.")
	default:
		log.Warn().Msgf("Unknown response type %T", r)
	}
}

// handle will take the response sent back by the Home Assistant API and run
// appropriate actions. This includes recording registration or setting disabled
// status.
func (t *SensorTracker) handle(response apiResponse, sensorUpdate Sensor) {
	log.Debug().
		Str("name", sensorUpdate.Name()).
		Str("id", sensorUpdate.ID()).
		Str("state", prettyPrintState(sensorUpdate)).
		Msg("Sensor updated.")
	if err := t.add(sensorUpdate); err != nil {
		log.Warn().Err(err).
			Str("name", sensorUpdate.Name()).
			Msg("Unable to add state for sensor to tracker.")
	}
	if response.Type() == api.ResponseTypeUpdate {
		switch {
		case response.Disabled():
			if err := t.registry.SetDisabled(sensorUpdate.ID(), true); err != nil {
				log.Warn().Err(err).
					Str("name", sensorUpdate.Name()).
					Msg("Unable to set as disabled in registry.")
			} else {
				log.Debug().
					Str("name", sensorUpdate.Name()).
					Msg("Sensor set to disabled.")
			}
		case !response.Disabled() && <-t.registry.IsDisabled(sensorUpdate.ID()):
			if err := t.registry.SetDisabled(sensorUpdate.ID(), false); err != nil {
				log.Warn().Err(err).
					Str("name", sensorUpdate.Name()).
					Msg("Unable to set as not disabled in registry.")
			}
		}
	}
	if response.Type() == api.ResponseTypeRegistration && response.Registered() {
		if err := t.registry.SetRegistered(sensorUpdate.ID(), true); err != nil {
			log.Warn().Err(err).
				Str("name", sensorUpdate.Name()).
				Msg("Unable to set as registered in registry.")
		} else {
			log.Debug().
				Str("name", sensorUpdate.Name()).
				Str("id", sensorUpdate.ID()).
				Msg("Sensor registered in Home Assistant.")
		}
	}
}

// UpdateSensors is the externally exposed method that devices can use to send a
// sensor state update.  It takes any number of sensor state updates of any type
// and handles them as appropriate.
func (t *SensorTracker) UpdateSensors(ctx context.Context, s any) {
	switch sensor := s.(type) {
	case Sensor:
		t.send(ctx, sensor)
	case *hass.LocationData:
		updateLocation(ctx, sensor)
	default:
		log.Warn().Msgf("Unknown sensor received %v", sensor)
	}
}

func (t *SensorTracker) Reset() {
	var err error
	if err = os.RemoveAll(t.registry.Path()); err != nil {
		log.Warn().Err(err).Msg("Could not remove existing registry DB.")
	}
	if t.registry, err = registry.NewJSONFilesRegistry(t.registry.Path()); err != nil {
		log.Warn().Err(err).Msg("Could not recreate registry.")
	}
	t.sensor = nil
}

func NewSensorTracker(id string) (*SensorTracker, error) {
	registryPath := filepath.Join(basePath, id, "sensorRegistry")
	db, err := registry.NewJSONFilesRegistry(registryPath)
	if err != nil {
		return nil, err
	}
	sensorTracker := &SensorTracker{
		registry: db,
		sensor:   make(map[string]Sensor),
	}
	return sensorTracker, nil
}
