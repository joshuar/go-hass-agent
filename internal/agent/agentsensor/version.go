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

type VersionWorker struct {
	prefs *preferences.CommonWorkerPrefs
}

func (w *VersionWorker) PreferencesID() string {
	return "version_sensor"
}

func (w *VersionWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *VersionWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
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

func NewVersionWorker(_ context.Context) *VersionWorker {
	worker := &VersionWorker{}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil
	}

	worker.prefs = prefs

	return worker
}
