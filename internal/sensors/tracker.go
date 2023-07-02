// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"errors"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type SensorTracker struct {
	registry   Registry
	sensor     map[string]*sensorState
	hassConfig *hass.HassConfig
	mu         sync.RWMutex
}

func NewSensorTracker(ctx context.Context, registryPath fyne.URI) *SensorTracker {
	r := &nutsdbRegistry{}
	err := r.Open(ctx, registryPath)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Unable to open registry")
		return nil
	}
	return &SensorTracker{
		registry:   r,
		sensor:     make(map[string]*sensorState),
		hassConfig: hass.NewHassConfig(ctx),
	}
}

// Add creates a new sensor in the tracker based on a recieved state
// update.
func (tracker *SensorTracker) add(sensor *sensorState) (*sensorState, error) {
	tracker.mu.Lock()
	if tracker.sensor == nil {
		tracker.mu.Unlock()
		return nil, errors.New("sensor map not initialised")
	}
	registryItem, err := tracker.registry.Get(sensor.ID())
	if err != nil {
		log.Debug().Caller().
			Msgf("Sensor %s not found in registry.", sensor.Name())
	}
	sensor.metadata.Registered = registryItem.data.Registered
	tracker.sensor[sensor.ID()] = sensor
	tracker.mu.Unlock()
	return sensor, nil
}

func (tracker *SensorTracker) isDisabled(id string) bool {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if tracker.sensor[id] != nil {
		return tracker.sensor[id].metadata.Disabled
	} else {
		return false
	}
}

func (tracker *SensorTracker) isRegistered(id string) bool {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if tracker.sensor[id] != nil {
		return tracker.sensor[id].metadata.Registered
	} else {
		return false
	}
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

func (tracker *SensorTracker) exists(id string) bool {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	if _, ok := tracker.sensor[id]; ok {
		return true
	} else {
		return false
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
func (tracker *SensorTracker) Update(ctx context.Context, s hass.SensorUpdate) {
	tracker.hassConfig.Refresh(ctx)
	// Assemble a sensor from the provided sensorUpdate, the HA config (for
	// disabled status) and the registry (for registered status)
	sensorID := s.ID()
	sensor := &sensorState{
		data: s,
		metadata: &sensorMetadata{
			Disabled: tracker.hassConfig.IsEntityDisabled(sensorID),
		},
	}
	// Update the registry with the latest assembled sensor data
	if sensor, err := tracker.add(sensor); err != nil {
		log.Debug().Caller().Err(err).
			Msg("Add sensor failed.")
	} else {
		// If the sensor is not disabled, update HA
		if !sensor.metadata.Disabled {
			hass.APIRequest(ctx, sensor)
		}
	}
}
