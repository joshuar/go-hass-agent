// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	registry "github.com/joshuar/go-hass-agent/internal/tracker/registry/jsonFiles"
	"github.com/rs/zerolog/log"
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

//go:generate moq -out mock_agentConfig_test.go . agentConfig
type agentConfig interface {
	GetConfig(string, interface{}) error
	StoragePath(string) (string, error)
}

type SensorTracker struct {
	registry Registry
	sensor   map[string]Sensor
	mu       sync.RWMutex
}

func RunSensorTracker(ctx context.Context, config agentConfig, trackerCh chan *SensorTracker) {
	registryPath, err := config.StoragePath(registryStorageID)
	if err != nil {
		log.Warn().Err(err).
			Msg("Path for sensor registry is not valid, using in-memory registry.")
	}
	db, err := registry.NewJsonFilesRegistry(registryPath)
	if err != nil {
		log.Error().Err(err).Msg("Unable to create a sensor tracker.")
		close(trackerCh)
	}
	sensorTracker := &SensorTracker{
		registry: db,
		sensor:   make(map[string]Sensor),
	}
	trackerCh <- sensorTracker
	var wg sync.WaitGroup
	updateCh := make(chan interface{})
	defer close(updateCh)
	sensorWorkers := device.SensorWorkers()
	sensorWorkers = append(sensorWorkers, device.ExternalIPUpdater)

	wg.Add(1)
	go func() {
		startWorkers(ctx, sensorWorkers, updateCh)
	}()
	wg.Add(1)
	go func() {
		sensorTracker.trackUpdates(ctx, config, updateCh)
	}()
	wg.Wait()
	close(trackerCh)
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
func (t *SensorTracker) updateSensor(ctx context.Context, config agentConfig, sensorUpdate Sensor) {
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
				Msgf("Failed to send sensor %s data to Home Assistant", sensorUpdate.Name())
		} else {
			log.Debug().
				Msgf("Sensor %s updated (%s). State is now: %v %s",
					sensorUpdate.Name(),
					sensorUpdate.ID(),
					sensorUpdate.State(),
					sensorUpdate.Units())
			if err := t.add(sensorUpdate); err != nil {
				log.Warn().Err(err).
					Msgf("Unable to add state for sensor %s to tracker.", sensorUpdate.Name())
			}
			if response.Type() == api.RequestTypeUpdateSensorStates {
				switch {
				case response.Disabled():
					if err := t.registry.SetDisabled(sensorUpdate.ID(), true); err != nil {
						log.Warn().Err(err).Msgf("Unable to set %s as disabled in registry.", sensorUpdate.Name())
					} else {
						log.Debug().Msgf("Sensor %s set to disabled.", sensorUpdate.Name())
					}
				case !response.Disabled() && <-t.registry.IsDisabled(sensorUpdate.ID()):
					if err := t.registry.SetDisabled(sensorUpdate.ID(), false); err != nil {
						log.Warn().Err(err).Msgf("Unable to set %s as not disabled in registry.", sensorUpdate.Name())
					}
				}
			}
			if response.Type() == api.RequestTypeRegisterSensor && response.Registered() {
				if err := t.registry.SetRegistered(sensorUpdate.ID(), true); err != nil {
					log.Warn().Err(err).Msgf("Unable to set %s as registered in registry.", sensorUpdate.Name())
				} else {
					log.Debug().Msgf("Sensor %s (%s) registered in Home Assistant.", sensorUpdate.Name(), sensorUpdate.ID())
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

func (t *SensorTracker) trackUpdates(ctx context.Context, config agentConfig, updateCh chan interface{}) {
	for {
		select {
		case data := <-updateCh:
			switch data := data.(type) {
			case Sensor:
				go t.updateSensor(ctx, config, data)
			case Location:
				go updateLocation(ctx, config, data)
			default:
				log.Warn().
					Msgf("Got unexpected status update %v", data)
			}
		case <-ctx.Done():
			log.Debug().
				Msg("Stopping sensor tracking.")
			return
		}
	}
}

// startWorkers will call all the sensor worker functions that have been defined
// for this device.
func startWorkers(ctx context.Context, workers []func(context.Context, chan interface{}), updateCh chan interface{}) {
	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go func(worker func(context.Context, chan interface{})) {
			defer wg.Done()
			// worker(workerCtx, updateCh)
			worker(ctx, updateCh)
		}(worker)
	}
	wg.Wait()
}
