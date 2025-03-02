// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package tracker

import (
	"errors"
	"sort"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/models"
)

var (
	ErrTrackerNotReady = errors.New("tracker not ready")
	ErrSensorNotFound  = errors.New("sensor not found in tracker")
)

type Tracker struct {
	sensor map[models.UniqueID]*models.Sensor
	mu     sync.Mutex
}

// Get fetches a sensors current tracked state.
func (t *Tracker) Get(id models.UniqueID) (*models.Sensor, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor[id] != nil {
		return t.sensor[id], nil
	}

	return nil, ErrSensorNotFound
}

func (t *Tracker) SensorList() []models.UniqueID {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor == nil {
		return nil
	}

	sortedEntities := make([]models.UniqueID, 0, len(t.sensor))

	for name := range t.sensor {
		sortedEntities = append(sortedEntities, name)
	}

	sort.Strings(sortedEntities)

	return sortedEntities
}

// Add creates a new sensor in the tracker based on a received state update.
func (t *Tracker) Add(details *models.Sensor) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sensor == nil {
		return ErrTrackerNotReady
	}

	t.sensor[details.UniqueID] = details

	return nil
}

func (t *Tracker) Reset() {
	if t.sensor != nil {
		t.sensor = nil
	}
}

func NewTracker() *Tracker {
	return &Tracker{
		sensor: make(map[models.UniqueID]*models.Sensor),
		mu:     sync.Mutex{},
	}
}
