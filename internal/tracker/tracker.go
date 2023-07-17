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
	"github.com/rs/zerolog/log"
)

type SensorTracker struct {
	registry Registry
	sensor   map[string]Sensor
	mu       sync.RWMutex
}

func RunSensorTracker(ctx context.Context, config api.Config) error {
	registryPath, err := config.NewStorage("sensorRegistry")
	if err != nil {
		log.Warn().Err(err).
			Msg("Path for sensor registry is not valid, using in-memory registry.")
	}
	db, err := NewNutsDB(ctx, registryPath)
	if err != nil {
		log.Error().Err(err).Msg("Unable to create a sensor tracker.")
		return err
		// return
	}
	sensorTracker := &SensorTracker{
		registry: db,
		sensor:   make(map[string]Sensor),
	}

	var wg sync.WaitGroup
	updateCh := make(chan interface{})
	defer close(updateCh)
	wg.Add(1)
	go func() {
		startWorkers(ctx, updateCh)
	}()
	// Sensors are tracked in a map to handle registration and
	// disabling/enabling. Updates are sent to Home Assistant.
	wg.Add(1)
	go func() {
		trackUpdates(ctx, sensorTracker, config, updateCh)
	}()
	wg.Wait()
	return nil
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

// updateSensor will send a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (tracker *SensorTracker) updateSensor(ctx context.Context, config api.Config, sensorUpdate Sensor) {
	// hassConfig, err := hass.GetHassConfig(ctx, config)
	// if err != nil {
	// 	log.Warn().Err(err).
	// 		Msg("Unable to retrieve config from Home Assistant.")
	// 	return
	// }
	// isDisabled, err := hassConfig.IsEntityDisabled(sensorUpdate.ID())
	// if err != nil {
	// 	log.Warn().Err(err).
	// 		Msgf("Unable to check disabled state for sensor %s in Home Assistant.", sensorUpdate.ID())
	// 	return
	// }
	// if isDisabled {
	// 	return
	// }
	var wg sync.WaitGroup
	var req api.Request
	if tracker.registry.IsRegistered(sensorUpdate.ID()) {
		req = marshalSensorUpdate(sensorUpdate)
	} else {
		req = marshalSensorRegistration(sensorUpdate)
	}
	responseCh := make(chan api.Response, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// defer close(responseCh)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, req, config, responseCh)
	}()
}

// startWorkers will call all the sensor worker functions that have been defined
// for this device.
func startWorkers(ctx context.Context, updateCh chan interface{}) {
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

func trackUpdates(ctx context.Context, tracker *SensorTracker, config api.Config, updateCh chan interface{}) {
	for {
		select {
		case data := <-updateCh:
			switch data := data.(type) {
			case Sensor:
				go tracker.updateSensor(ctx, config, data)
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
