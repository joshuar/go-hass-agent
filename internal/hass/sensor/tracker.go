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
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

var (
	ErrRespFailed      = errors.New("unsuccessful request")
	ErrRespUnknown     = errors.New("unhandled response")
	ErrTrackerNotReady = errors.New("tracker not ready")
	ErrSensorNotFound  = errors.New("sensor not found in tracker")
)

//go:generate moq -out mock_Registry_test.go . Registry
type Registry interface {
	SetDisabled(sensor string, state bool) error
	SetRegistered(sensor string, state bool) error
	IsDisabled(sensor string) bool
	IsRegistered(sensor string) bool
}

type Tracker struct {
	sensor map[string]Details
	mu     sync.Mutex
}

// Get fetches a sensors current tracked state.
func (t *Tracker) Get(id string) (Details, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor[id] != nil {
		return t.sensor[id], nil
	}

	return nil, ErrSensorNotFound
}

func (t *Tracker) SensorList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor == nil {
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

func (t *Tracker) Process(ctx context.Context, reg Registry, upds ...<-chan Details) error {
	client := hass.ContextGetClient(ctx)

	for update := range MergeSensorCh(ctx, upds...) {
		go func(upd Details) {
			if err := t.update(ctx, client, "", reg, upd); err != nil {
				log.Warn().Err(err).Str("id", upd.ID()).Msg("Update failed.")
			} else {
				log.Debug().
					Str("name", upd.Name()).
					Str("id", upd.ID()).
					Interface("state", upd.State()).
					Str("units", upd.Units()).
					Msg("Sensor updated.")
			}
		}(update)
	}

	return nil
}

// Add creates a new sensor in the tracker based on a received state update.
func (t *Tracker) add(sensor Details) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor == nil {
		return ErrTrackerNotReady
	}

	t.sensor[sensor.ID()] = sensor

	return nil
}

// update will update a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (t *Tracker) update(ctx context.Context, client *resty.Client, url string, reg Registry, upd Details) error {
	req, resp, err := NewRequest(reg, upd)
	if err != nil {
		return wrapErr(upd.ID(), err)
	}
	// Send the sensor request to Home Assistant.
	if err := hass.ExecuteRequest(ctx, client, url, req, resp); err != nil {
		return wrapErr(upd.ID(), err)
	}
	// Handle the response received.
	if err := handleResponse(resp, t, upd, reg); err != nil {
		return wrapErr(upd.ID(), err)
	}

	return nil
}

func handleResponse(respIntr hass.Response, trk *Tracker, upd Details, reg Registry) error {
	switch resp := respIntr.(type) {
	case *updateResponse:
		if err := handleUpdates(reg, resp); err != nil {
			return err
		}

		if err := trk.add(upd); err != nil {
			return err
		}
	case *registrationResponse:
		if err := handleRegistration(reg, resp, upd.ID()); err != nil {
			return err
		}
	case *locationResponse:
		return nil
	default:
		return ErrRespUnknown
	}

	return nil
}

//nolint:err113
func handleUpdates(reg Registry, r *updateResponse) error {
	for sensor, details := range r.Body {
		if details == nil {
			return ErrRespUnknown
		}

		if !details.Success {
			if details.Error != nil {
				return fmt.Errorf("%d: %s", details.Error.Code, details.Error.Message)
			}

			return ErrRespFailed
		}

		if reg.IsDisabled(sensor) != details.Disabled {
			if err := reg.SetDisabled(sensor, details.Disabled); err != nil {
				return fmt.Errorf("could not set disabled status: %w", err)
			}
		}
	}

	return nil
}

func handleRegistration(reg Registry, r *registrationResponse, s string) error {
	if !r.Body.Success {
		return ErrRespFailed
	}

	err := reg.SetRegistered(s, true)
	if err != nil {
		return fmt.Errorf("could not register: %w", err)
	}

	return nil
}

func (t *Tracker) Reset() {
	t.sensor = nil
}

func NewTracker() (*Tracker, error) {
	sensorTracker := &Tracker{
		sensor: make(map[string]Details),
		mu:     sync.Mutex{},
	}

	return sensorTracker, nil
}

func MergeSensorCh(ctx context.Context, sensorCh ...<-chan Details) chan Details {
	var wg sync.WaitGroup

	out := make(chan Details)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(sensorOutCh <-chan Details) {
		defer wg.Done()

		if sensorOutCh == nil {
			return
		}

		for n := range sensorOutCh {
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

func wrapErr(sensorID string, e error) error {
	return fmt.Errorf("%s update failed: %w", sensorID, e)
}
