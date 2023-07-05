// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"errors"
	"sync"

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

func NewSensorTracker(ctx context.Context, path string) *SensorTracker {
	r := &nutsdbRegistry{}
	err := r.Open(ctx, path)
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
func (tracker *SensorTracker) add(sensor *sensorState) error {
	tracker.mu.Lock()
	if tracker.sensor == nil {
		tracker.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	tracker.sensor[sensor.ID()] = sensor
	tracker.mu.Unlock()
	tracker.registry.Set(RegistryItem{
		data: sensor.metadata,
		id:   sensor.ID(),
	})
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
func (tracker *SensorTracker) Update(ctx context.Context, s hass.SensorUpdate) {
	// Assemble a sensor from the provided sensorUpdate, the HA config (for
	// disabled status) and the registry (for registered status)
	metadata, err := tracker.registry.Get(s.ID())
	if err != nil {
		log.Debug().Err(err).Msg("Error getting tracker metadata from registry.")
		metadata = NewRegistryItem(s.ID())
	}
	if err := tracker.hassConfig.Refresh(ctx); err != nil {
		log.Debug().Err(err).
			Msg("Unable to fetch updated config from Home Assistant")
	}
	metadata.data.Disabled = tracker.hassConfig.IsEntityDisabled(s.ID())
	sensor := &sensorState{
		data:       s,
		metadata:   metadata.data,
		DisabledCh: make(chan bool, 1),
	}
	if !sensor.Disabled() {
		go hass.APIRequest(ctx, sensor)
		sensor.metadata.Disabled = <-sensor.DisabledCh
		if err := tracker.add(sensor); err != nil {
			log.Debug().Err(err).Msgf("Error adding sensor %s to registry.", s.ID())
		}
	}
}
