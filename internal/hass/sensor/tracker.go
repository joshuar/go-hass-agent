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

	"github.com/joshuar/go-hass-agent/internal/hass"
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

// Add creates a new sensor in the tracker based on a received state update.
func (t *Tracker) add(s Details) error {
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
func (t *Tracker) Get(id string) (Details, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sensor[id] != nil {
		return t.sensor[id], nil
	} else {
		return nil, errors.New("not found")
	}
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

// UpdateSensor will UpdateSensor a sensor update to HA, checking to ensure the sensor is not
// disabled. It will also update the local registry state based on the response.
func (t *Tracker) UpdateSensor(ctx context.Context, reg Registry, upd Details) error {
	if reg.IsDisabled(upd.ID()) {
		return nil
	}
	var req hass.PostRequest
	var resp hass.Response
	if reg.IsRegistered(upd.ID()) {
		req = NewUpdateRequest(SensorState(upd))
		resp = NewUpdateResponse()
	} else {
		req = NewRegistrationRequest(SensorRegistration(upd))
		resp = NewRegistrationResponse()
	}
	hass.ExecuteRequest(ctx, req, resp)
	if errors.Is(resp, &hass.APIError{}) || resp.Error() != "" {
		return wrapErr(upd.ID(), resp)
	}
	if err := handleResponse(resp, t, upd, reg); err != nil {
		return wrapErr(upd.ID(), err)
	}
	return nil
}

func handleResponse(resp hass.Response, trk *Tracker, upd Details, reg Registry) error {
	switch r := resp.(type) {
	case *updateResponse:
		if err := handleUpdates(reg, r); err != nil {
			return err
		}
		if err := trk.add(upd); err != nil {
			return err
		}
	case *registrationResponse:
		if err := handleRegistration(reg, r, upd.ID()); err != nil {
			return err
		}
	default:
		return errors.New("unhandled response")
	}
	return nil
}

func handleUpdates(reg Registry, r *updateResponse) error {
	for sensor, details := range r.Body {
		if details == nil {
			return errors.New("empty response")
		}
		if !details.Success {
			if details.Error != nil {
				return fmt.Errorf("%d: %s", details.Error.Code, details.Error.Message)
			}
			return errors.New("update unsuccessful")
		}
		if reg.IsDisabled(sensor) != details.Disabled {
			if err := reg.SetDisabled(sensor, details.Disabled); err != nil {
				return fmt.Errorf("could no set disabled status: %w", err)
			}
		}
	}
	return nil
}

func handleRegistration(reg Registry, r *registrationResponse, s string) error {
	if !r.Body.Success {
		return errors.New("registration unsuccessful")
	}
	return reg.SetRegistered(s, true)
}

func (t *Tracker) Reset() {
	t.sensor = nil
}

func NewTracker() (*Tracker, error) {
	sensorTracker := &Tracker{
		sensor: make(map[string]Details),
	}
	return sensorTracker, nil
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

func wrapErr(sensorID string, e error) error {
	return fmt.Errorf("%s update failed: %w", sensorID, e)
}
