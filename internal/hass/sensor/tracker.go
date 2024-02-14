// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

//go:generate moq -out mock_Registry_test.go . Registry
type Registry interface {
	SetDisabled(sensor string, state bool) error
	SetRegistered(sensor string, state bool) error
	IsDisabled(sensor string) bool
	IsRegistered(sensor string) bool
}

type SensorTracker struct {
	sensor map[string]Details
	mu     sync.Mutex
}

// Add creates a new sensor in the tracker based on a received state update.
func (t *SensorTracker) add(s Details) error {
	t.mu.Lock()
	if t.sensor == nil {
		t.mu.Unlock()
		return errors.New("sensor map not initialised")
	}
	t.sensor[s.ID()] = s
	t.mu.Unlock()
	return nil
}

// Get fetches a sensors current tracked state.
func (t *SensorTracker) Get(id string) (Details, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sensor[id] != nil {
		return t.sensor[id], nil
	} else {
		return nil, errors.New("not found")
	}
}

func (t *SensorTracker) SensorList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sensor == nil {
		log.Warn().Msg("No sensors available.")
		return nil
	}
	sortedEntities := make([]string, 0, len(t.sensor))
	for name, sensor := range t.sensor {
		if sensor.State() != nil {
			sortedEntities = append(sortedEntities, name)
		}
	}
	sort.Strings(sortedEntities)
	return sortedEntities
}

// UpdateSensor will UpdateSensor a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (t *SensorTracker) UpdateSensor(ctx context.Context, reg Registry, upd Details) {
	if reg.IsDisabled(upd.ID()) {
		log.Debug().Str("id", upd.ID()).
			Msg("Sensor is disabled. Ignoring update.")
		return
	}
	var r any
	if reg.IsRegistered(upd.ID()) {
		r = UpdateRequest(SensorState(upd))
	} else {
		r = RegistrationRequest(SensorRegistration(upd))
	}
	resp := <-hass.ExecuteRequest(ctx, r)
	if resp.Error != nil {
		err := fmt.Errorf("%d: %s", resp.Error.StatusCode, resp.Error.Message)
		log.Warn().Err(err).Str("id", upd.ID()).
			Msg("Failed to send sensor data to Home Assistant.")
	}
	switch r := resp.Body.(type) {
	case *UpdateResponse:
		if err := t.handleUpdates(reg, r); err != nil {
			log.Warn().Err(err).Msg("Sensor update unsuccessful.")
		} else {
			log.Debug().
				Str("name", upd.Name()).
				Str("id", upd.ID()).
				Str("state", prettyPrintState(upd)).
				Msg("Sensor updated.")
			if err := t.add(upd); err != nil {
				log.Warn().Err(err).
					Str("name", upd.Name()).
					Msg("Unable to add state for sensor to tracker.")
			}
		}
	case *RegistrationResponse:
		if err := t.handleRegistration(reg, r, upd.ID()); err != nil {
			log.Warn().Err(err).Str("id", upd.Name()).Msg("Unable to register ")
		}
		log.Debug().
			Str("name", upd.Name()).
			Str("id", upd.ID()).
			Msg("Sensor registered.")
	default:
		log.Warn().Interface("response", r).
			Str("id", upd.ID()).
			Msg("Unhandled response from Home Assistant.")
	}
}

func (t *SensorTracker) handleUpdates(reg Registry, r *UpdateResponse) error {
	for sensor, details := range *r {
		if !details.Success {
			return fmt.Errorf("%d: %s", details.Error.Code, details.Error.Message)
		}
		if reg.IsDisabled(sensor) != details.Disabled {
			if err := reg.SetDisabled(sensor, details.Disabled); err != nil {
				log.Warn().Err(err).Str("id", sensor).Msg("Could not set disabled status in registry.")
			} else {
				log.Info().Str("id", sensor).Msg("Sensor disabled.")
			}
		}
	}
	return nil
}

func (t *SensorTracker) handleRegistration(reg Registry, r *RegistrationResponse, s string) error {
	if !r.Success {
		return errors.New("registration unsuccessful")
	}
	return reg.SetRegistered(s, true)
}

func (t *SensorTracker) Reset() {
	t.sensor = nil
}

func NewSensorTracker() (*SensorTracker, error) {
	sensorTracker := &SensorTracker{
		sensor: make(map[string]Details),
	}
	return sensorTracker, nil
}

func prettyPrintState(s Details) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v", s.State())
	if s.Units() != "" {
		fmt.Fprintf(&b, " %s", s.Units())
	}
	return b.String()
}

func MergeSensorCh(ctx context.Context, sensorCh ...<-chan Details) chan Details {
	var wg sync.WaitGroup
	out := make(chan Details)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan Details) {
		defer wg.Done()
		if c == nil {
			return
		}
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(sensorCh))
	for _, c := range sensorCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
