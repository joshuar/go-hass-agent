// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	versionWorkerID = "agent_version_sensor"
)

func newVersionSensor() sensor.Entity {
	return sensor.Entity{
		Name:     "Go Hass Agent Version",
		Category: types.CategoryDiagnostic,
		EntityState: &sensor.EntityState{
			ID:         "agent_version",
			Icon:       "mdi:face-agent",
			EntityType: types.Sensor,
			State:      preferences.AppVersion,
		},
	}
}

type versionWorker struct{}

func (w *versionWorker) ID() string { return versionWorkerID }

func (w *versionWorker) Stop() error { return nil }

func (w *versionWorker) Start(_ context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	go func() {
		defer close(sensorCh)
		sensorCh <- newVersionSensor()
	}()

	return sensorCh, nil
}

func (w *versionWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return []sensor.Entity{newVersionSensor()}, nil
}

func newVersionWorker() *versionWorker {
	return &versionWorker{}
}
