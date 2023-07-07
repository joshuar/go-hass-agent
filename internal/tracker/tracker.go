// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"errors"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/request"
	"github.com/rs/zerolog/log"
)

type SensorTracker struct {
	registry Registry
	sensor   map[string]*sensorState
	mu       sync.RWMutex
}

func NewSensorTracker(ctx context.Context, path string) *SensorTracker {
	db, err := NewNutsDB(ctx, path)
	if err != nil {
		log.Debug().Err(err).
			Msg("Could not open database.")
		return nil
	}
	return &SensorTracker{
		registry: db,
		sensor:   make(map[string]*sensorState),
	}
}

// Add creates a new sensor in the tracker based on a recieved state
// update.
func (tracker *SensorTracker) add(sensor *sensorState) error {
	tracker.mu.Lock()
	if tracker.sensor == nil {
		tracker.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	tracker.sensor[sensor.data.ID()] = sensor
	tracker.mu.Unlock()
	return nil
}

// Get fetches a sensors current tracked state
func (tracker *SensorTracker) Get(id string) (sensorState, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if tracker.sensor[id] != nil {
		return *tracker.sensor[id], nil
	} else {
		return sensorState{}, errors.New("not found")
	}
}

// StartWorkers will call all the sensor worker functions that have been defined
// for this device.
func (tracker *SensorTracker) StartWorkers(ctx context.Context, updateCh chan interface{}) {
	var wg sync.WaitGroup

	// Run all the defined sensor update functions.
	deviceAPI, err := device.FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
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
	// Assemble a sensor from the provided sensorUpdate, the HA config (for
	// disabled status) and the registry (for registered status)
	// metadata, err := tracker.registry.Get(sensorUpdate.ID())
	// if err != nil {
	// 	log.Debug().Err(err).Msg("Error getting tracker metadata from registry.")
	// 	metadata = NewRegistryItem(sensorUpdate.ID())
	// }
	hassConfig := hass.NewHassConfig(ctx)
	if hassConfig == nil {
		log.Debug().
			Msg("Unable to fetch updated config from Home Assistant.")
	}
	if hassConfig.IsEntityDisabled(sensorUpdate.ID()) {
		log.Debug().Msgf("Received sensor update for disabled sensor %s.", sensorUpdate.ID())
	} else {
		var wg sync.WaitGroup
		sensor := newSensorState(sensorUpdate, tracker.registry)
		wg.Add(1)
		go func() {
			defer wg.Done()
			request.APIRequest(ctx, sensor)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.registry.SetDisabled(sensorUpdate.ID(), <-sensor.disableCh)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := <-sensor.errCh; err != nil {
				log.Debug().Err(err).
					Msgf("Failed to send sensor %s data to Home Assistant", sensorUpdate.Name())
			} else {
				if !tracker.registry.IsRegistered(sensorUpdate.ID()) {
					tracker.registry.SetRegistered(sensorUpdate.ID(), true)
					log.Debug().Caller().
						Msgf("Sensor %s registered in HA.",
							sensor.data.Name())
				} else {
					log.Debug().Caller().
						Msgf("Sensor %s updated (%s). State is now: %v %s",
							sensor.data.Name(),
							sensor.data.ID(),
							sensor.data.State(),
							sensor.data.Units())
				}
			}
		}()
		// if err := tracker.add(sensor); err != nil {
		// 	log.Debug().Err(err).Msgf("Error adding sensor %s to registry.", sensorUpdate.ID())
		// }
	}
}
