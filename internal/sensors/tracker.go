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
	registry Registry
	sensor   map[string]*sensorState
	mu       sync.RWMutex
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
		sensor:   make(map[string]*sensorState),
		registry: r,
	}
}

// Add creates a new sensor in the tracker based on a recieved state
// update.
func (tracker *SensorTracker) add(s hass.SensorUpdate) error {
	tracker.mu.Lock()
	if tracker.sensor == nil {
		tracker.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	state := marshalSensorState(s)
	registryItem, err := tracker.registry.Get(state.entityID)
	if err != nil {
		log.Debug().Caller().
			Msgf("Sensor %s not found in registry.", s.Name())
	}
	state.metadata = registryItem.data
	tracker.sensor[state.entityID] = state
	tracker.mu.Unlock()
	if tracker.exists(state.entityID) {
		log.Debug().Caller().Msgf("Sensor: %s added (%s).", state.name, state.entityID)
		return nil
	} else {
		return errors.New("sensor was not added")
	}
}

// Get fetches a sensors current tracked state
func (tracker *SensorTracker) Get(id string) *sensorState {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	return tracker.sensor[id]
}

func (tracker *SensorTracker) update(s hass.SensorUpdate) error {
	if !tracker.exists(s.ID()) {
		return errors.New("sensor not found")
	}
	tracker.mu.Lock()
	tracker.sensor[s.ID()].state = s.State()
	tracker.sensor[s.ID()].attributes = s.Attributes()
	tracker.sensor[s.ID()].icon = s.Icon()
	tracker.mu.Unlock()
	return nil
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
func (tracker *SensorTracker) Update(ctx context.Context, s hass.SensorUpdate, c *hass.HassConfig) {
	sensorID := s.ID()
	var err error
	if !tracker.exists(sensorID) {
		err = tracker.add(s)
	} else {
		err = tracker.update(s)
	}
	if err == nil {
		sensor := tracker.Get(sensorID)
		if c.IsEntityDisabled(sensorID) {
			if !sensor.metadata.Disabled {
				sensor.metadata.Disabled = true
			}
		} else {
			hass.APIRequest(ctx, sensor)
			tracker.registry.Set(registryItem{id: sensorID, data: sensor.metadata})
		}
	}
}
