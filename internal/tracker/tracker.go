// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	registry "github.com/joshuar/go-hass-agent/internal/tracker/registry/jsonFiles"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	registryStorageID = "sensorRegistry"
)

//go:generate moq -out mock_Registry_test.go . Registry
type Registry interface {
	SetDisabled(string, bool) error
	SetRegistered(string, bool) error
	IsDisabled(string) chan bool
	IsRegistered(string) chan bool
}

//go:generate moq -out mock_agent_test.go . agent
type agent interface {
	GetConfig(string, interface{}) error
	StoragePath(string) (string, error)
}

type SensorTracker struct {
	registry    Registry
	agentConfig agent
	sensor      map[string]Sensor
	mu          sync.RWMutex
}

// Add creates a new sensor in the tracker based on a recieved state update.
func (tracker *SensorTracker) add(s Sensor) error {
	tracker.mu.Lock()
	if tracker.sensor == nil {
		tracker.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	tracker.sensor[s.ID()] = s
	tracker.mu.Unlock()
	return nil
}

// Get fetches a sensors current tracked state
func (tracker *SensorTracker) Get(id string) (Sensor, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if tracker.sensor[id] != nil {
		return tracker.sensor[id], nil
	} else {
		return nil, errors.New("not found")
	}
}

func (tracker *SensorTracker) SensorList() []string {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if tracker.sensor == nil {
		log.Warn().Msg("No sensors available.")
		return nil
	}
	sortedEntities := make([]string, 0, len(tracker.sensor))
	for name, sensor := range tracker.sensor {
		if sensor.State() != nil {
			sortedEntities = append(sortedEntities, name)
		}
	}
	sort.Strings(sortedEntities)
	return sortedEntities
}

// updateSensor will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (t *SensorTracker) updateSensor(ctx context.Context, config agent, sensorUpdate Sensor) {
	var wg sync.WaitGroup
	var req api.Request
	if disabled := <-t.registry.IsDisabled(sensorUpdate.ID()); disabled {
		log.Debug().Msgf("Sensor %s is disabled. Ignoring update.", sensorUpdate.ID())
	}
	registered := <-t.registry.IsRegistered(sensorUpdate.ID())
	switch registered {
	case true:
		req = marshalSensorUpdate(sensorUpdate)
	case false:
		req = marshalSensorRegistration(sensorUpdate)
	}
	responseCh := make(chan api.Response, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		response := <-responseCh
		if response.Error() != nil {
			log.Error().Err(response.Error()).
				Str("name", sensorUpdate.Name()).
				Msg("Failed to send sensor data to Home Assistant.")
		} else {
			log.Debug().
				Str("name", sensorUpdate.Name()).
				Str("id", sensorUpdate.ID()).
				Str("state", fmt.Sprintf("%v %s", sensorUpdate.State(), sensorUpdate.Units())).
				Msg("Sensor updated.")
			if err := t.add(sensorUpdate); err != nil {
				log.Warn().Err(err).
					Str("name", sensorUpdate.Name()).
					Msg("Unable to add state for sensor to tracker.")
			}
			if response.Type() == api.RequestTypeUpdateSensorStates {
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
			if response.Type() == api.RequestTypeRegisterSensor && response.Registered() {
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
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, req, config, responseCh)
	}()
	wg.Wait()
}

// UpdateSensors is the externally exposed method that devices can use to send a
// sensor state update.  It takes any number of sensor state updates of any type
// and handles them as appropriate.
func (t *SensorTracker) UpdateSensors(ctx context.Context, sensors ...interface{}) error {
	g, _ := errgroup.WithContext(ctx)
	sensorData := make(chan interface{}, len(sensors))

	for i := 0; i < len(sensors); i++ {
		sensorData <- sensors[i]
	}

	g.Go(func() error {
		var i int
		for s := range sensorData {
			switch sensor := s.(type) {
			case Sensor:
				t.updateSensor(ctx, t.agentConfig, sensor)
			case Location:
				updateLocation(ctx, t.agentConfig, sensor)
			}
			i++
		}
		log.Trace().Int("sensorsUpdated", i).Msg("Finished updating sensors.")
		return nil
	})

	close(sensorData)
	return g.Wait()
}

func NewSensorTracker(agentConfig agent) (*SensorTracker, error) {
	registryPath, err := agentConfig.StoragePath(registryStorageID)
	if err != nil {
		log.Warn().Err(err).
			Msg("Path for sensor registry is not valid, using in-memory registry.")
	}
	db, err := registry.NewJsonFilesRegistry(registryPath)
	if err != nil {
		return nil, errors.New("unable to create a sensor tracker")
	}
	sensorTracker := &SensorTracker{
		registry:    db,
		sensor:      make(map[string]Sensor),
		agentConfig: agentConfig,
	}
	return sensorTracker, nil
}
