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

func (tracker *sensorTracker) Add(s hass.SensorUpdate) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.newState(s)
}

func (tracker *sensorTracker) Update(s hass.SensorUpdate) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.updateState(s)
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

// func (tracker *sensorTracker) Disabled(id string) bool {
// 	return tracker.sensor[id].Disabled
// }

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

func (tracker *sensorTracker) newState(newSensor hass.SensorUpdate) {
	sensor := &sensorState{
		entityID:    newSensor.ID(),
		name:        newSensor.Name(),
		deviceClass: newSensor.DeviceClass(),
		stateClass:  newSensor.StateClass(),
		sensorType:  newSensor.SensorType(),
		state:       newSensor.State(),
		attributes:  newSensor.Attributes(),
		icon:        newSensor.Icon(),
		stateUnits:  newSensor.Units(),
		category:    newSensor.Category(),
		metadata:    tracker.registry.Add(newSensor.ID()),
	}
	tracker.sensor[newSensor.ID()] = sensor
}

// updateSensor ensures the bare minimum properties of a sensor are updated from
// a hass.SensorUpdate
func (tracker *sensorTracker) updateState(update hass.SensorUpdate) {
	tracker.sensor[update.ID()].state = update.State()
	tracker.sensor[update.ID()].attributes = update.Attributes()
	tracker.sensor[update.ID()].icon = update.Icon()
}

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
