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

type sensorTracker struct {
	mu            sync.RWMutex
	sensor        map[string]*sensorState
	sensorWorkers *device.SensorInfo
	registry      *sensorRegistry
	hassConfig    *hass.HassConfig
}

func NewSensorTracker(ctx context.Context, appPath fyne.URI) *sensorTracker {
	return &sensorTracker{
		sensor:        make(map[string]*sensorState),
		sensorWorkers: device.SetupSensors(),
		registry:      OpenSensorRegistry(ctx, appPath),
		hassConfig:    hass.NewHassConfig(ctx),
	}
}

// Add creates a new sensor in the tracker based on a recieved state
// update.
func (tracker *sensorTracker) add(s hass.SensorUpdate) error {
	tracker.mu.Lock()
	state := marshalSensorState(s)
	metadata, err := tracker.registry.Get(state.entityID)
	if err != nil {
		log.Debug().Caller().
			Msgf("Sensor %s not found in registry.", s.Name())
	}
	state.metadata = metadata
	tracker.sensor[state.entityID] = state
	tracker.mu.Unlock()
	if tracker.exists(state.entityID) {
		log.Debug().Caller().Msgf("Added sensor: %s", state.entityID)
		return nil
	} else {
		return errors.New("sensor was not added")
	}
}

// Update will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (tracker *sensorTracker) Update(ctx context.Context, s hass.SensorUpdate) {
	sensorID := s.ID()
	var err error
	if !tracker.exists(sensorID) {
		err = tracker.add(s)
	} else {
		tracker.update(s)
	}
	if err == nil {
		sensor := tracker.get(sensorID)
		if tracker.hassConfig.IsEntityDisabled(sensorID) {
			if !sensor.metadata.Disabled {
				sensor.metadata.Disabled = true
			}
		} else {
			hass.APIRequest(ctx, sensor)
			tracker.registry.Set(sensorID, sensor.metadata)
		}
	}
}

func (tracker *sensorTracker) get(id string) *sensorState {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()
	return tracker.sensor[id]
}

func (tracker *sensorTracker) update(s hass.SensorUpdate) {
	tracker.mu.Lock()
	tracker.sensor[s.ID()].state = s.State()
	tracker.sensor[s.ID()].attributes = s.Attributes()
	tracker.sensor[s.ID()].icon = s.Icon()
	tracker.mu.Unlock()
}

func (tracker *sensorTracker) exists(id string) bool {
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
func (tracker *sensorTracker) StartWorkers(ctx context.Context, updateCh chan interface{}) {
	var wg sync.WaitGroup

	// Run all the defined sensor update functions.
	for name, workerFunction := range tracker.sensorWorkers.Get() {
		wg.Add(1)
		log.Debug().Caller().
			Msgf("Setting up sensors for %s.", name)
		go func(worker func(context.Context, chan interface{})) {
			defer wg.Done()
			worker(ctx, updateCh)
		}(workerFunction)
	}
	wg.Wait()
}
