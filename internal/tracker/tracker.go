// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"errors"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type SensorTracker struct {
	registry Registry
	sensor   map[string]Sensor
	mu       sync.RWMutex
}

func NewSensorTracker(ctx context.Context, path string) *SensorTracker {
	db, err := NewNutsDB(ctx, path)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not open registry database.")
		return nil
	}
	return &SensorTracker{
		registry: db,
		sensor:   make(map[string]Sensor),
	}
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

// StartWorkers will call all the sensor worker functions that have been defined
// for this device.
func (tracker *SensorTracker) StartWorkers(ctx context.Context, updateCh chan interface{}) {
	var wg sync.WaitGroup

	// Run all the defined sensor update functions.
	deviceAPI, err := device.FetchAPIFromContext(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not fetch sensor workers.")
		return
	}
	sensorWorkers := deviceAPI.SensorWorkers()
	sensorWorkers = append(sensorWorkers, device.ExternalIPUpdater)
	for _, worker := range sensorWorkers {
		wg.Add(1)
		go func(worker func(context.Context, chan interface{})) {
			defer wg.Done()
			worker(ctx, updateCh)
		}(worker)
	}
	wg.Wait()
}

// Update will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (tracker *SensorTracker) Update(ctx context.Context, sensorUpdate Sensor) {
	isDisabled, err := hass.IsEntityDisabled(ctx, sensorUpdate.ID())
	if err != nil {
		log.Warn().Err(err).
			Msgf("Unable to check disabled state for sensor %s in Home Assistant.", sensorUpdate.ID())
		return
	}
	if isDisabled {
		return
	}
	var wg sync.WaitGroup
	var req api.Request
	if tracker.registry.IsRegistered(sensorUpdate.ID()) {
		req = MarshalSensorUpdate(sensorUpdate)
	} else {
		req = MarshalSensorRegistration(sensorUpdate)
	}
	responseCh := make(chan api.Response, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, req, responseCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(responseCh)
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
			if err := tracker.add(sensorUpdate); err != nil {
				log.Warn().Err(err).
					Msgf("Unable to add state for sensor %s to tracker.", sensorUpdate.Name())
			}
			if response.Type() == api.RequestTypeUpdateSensorStates && response.Disabled() {
				if err := tracker.registry.SetDisabled(sensorUpdate.ID(), true); err != nil {
					log.Warn().Err(err).Msgf("Unable to set %s as disabled in registry.", sensorUpdate.Name())
				}
			}
			if response.Type() == api.RequestTypeRegisterSensor && response.Registered() {
				log.Debug().Msgf("Sensor %s (%s) registered in Home Assistant.", sensorUpdate.Name(), sensorUpdate.ID())
				if err := tracker.registry.SetRegistered(sensorUpdate.ID(), true); err != nil {
					log.Warn().Err(err).Msgf("Unable to set %s as registered in registry.", sensorUpdate.Name())
				}
			}
		}
	}()
}
