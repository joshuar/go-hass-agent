// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type sensorTracker struct {
	mu            sync.Mutex
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
func (tracker *sensorTracker) Add(s hass.SensorUpdate) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	sensor := &sensorState{
		entityID:    s.ID(),
		name:        s.Name(),
		deviceClass: s.DeviceClass(),
		stateClass:  s.StateClass(),
		sensorType:  s.SensorType(),
		state:       s.State(),
		attributes:  s.Attributes(),
		icon:        s.Icon(),
		stateUnits:  s.Units(),
		category:    s.Category(),
		metadata:    tracker.registry.Add(s.ID()),
	}
	tracker.sensor[s.ID()] = sensor
}

// Update ensures the bare minimum properties of a sensor are updated from
// a hass.SensorUpdate
func (tracker *sensorTracker) Update(s hass.SensorUpdate) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.sensor[s.ID()].state = s.State()
	tracker.sensor[s.ID()].attributes = s.Attributes()
	tracker.sensor[s.ID()].icon = s.Icon()
}

func (tracker *sensorTracker) Get(id string) *sensorState {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return tracker.sensor[id]
}

func (tracker *sensorTracker) Exists(id string) bool {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	if _, ok := tracker.sensor[id]; ok {
		return true
	} else {
		return false
	}
}

// Send will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (tracker *sensorTracker) Send(ctx context.Context, id string) {
	if tracker.hassConfig.IsEntityDisabled(id) {
		if !tracker.sensor[id].metadata.IsDisabled() {
			tracker.sensor[id].metadata.SetDisabled(true)
		}
	} else {
		hass.APIRequest(ctx, tracker.sensor[id])
		tracker.registry.Update(tracker.sensor[id].metadata)
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
