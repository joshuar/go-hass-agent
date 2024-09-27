// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrTrackerNotReady = errors.New("tracker not ready")
	ErrSensorNotFound  = errors.New("sensor not found in tracker")
)

type Tracker struct {
	sensor map[string]*Entity
	mu     sync.Mutex
}

// Get fetches a sensors current tracked state.
func (t *Tracker) Get(id string) (*Entity, error) {
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

	for name := range t.sensor {
		sortedEntities = append(sortedEntities, name)
	}

	sort.Strings(sortedEntities)

	return sortedEntities
}

// Add creates a new sensor in the tracker based on a received state update.
func (t *Tracker) Add(sensor *Entity) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor == nil {
		return ErrTrackerNotReady
	}

	t.sensor[sensor.ID] = sensor

	return nil
}

func (t *Tracker) Reset() {
	if t.sensor != nil {
		t.sensor = nil
	}
}

func NewTracker() *Tracker {
	return &Tracker{
		sensor: make(map[string]*Entity),
		mu:     sync.Mutex{},
	}
}
