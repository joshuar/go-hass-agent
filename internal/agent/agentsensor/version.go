// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agentsensor

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	versionWorkerID = "agent_version"
)

func newVersionSensor() sensor.Entity {
	return sensor.NewSensor(
		sensor.WithName("Go Hass Agent Version"),
		sensor.WithID("agent_version"),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon("mdi:face-agent"),
			sensor.WithValue(preferences.AppVersion()),
		),
	)
}

type VersionWorker struct{}

// TODO: implement ability to disable.
func (w *VersionWorker) Disabled() bool {
	return false
}

func (w *VersionWorker) ID() string { return versionWorkerID }

func (w *VersionWorker) Stop() error { return nil }

func (w *VersionWorker) Start(_ context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	go func() {
		defer close(sensorCh)
		sensorCh <- newVersionSensor()
	}()

	return sensorCh, nil
}

func (w *VersionWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return []sensor.Entity{newVersionSensor()}, nil
}

func NewVersionWorker() *VersionWorker {
	return &VersionWorker{}
}
