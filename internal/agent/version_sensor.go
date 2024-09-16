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

type version string

func (v *version) Name() string { return "Go Hass Agent Version" }

func (v *version) ID() string { return "agent_version" }

func (v *version) Icon() string { return "mdi:face-agent" }

func (v *version) SensorType() types.SensorClass { return types.Sensor }

func (v *version) DeviceClass() types.DeviceClass { return 0 }

func (v *version) StateClass() types.StateClass { return 0 }

func (v *version) State() any { return preferences.AppVersion }

func (v *version) Units() string { return "" }

func (v *version) Category() string { return types.CategoryDiagnostic }

func (v *version) Attributes() map[string]any { return nil }

type versionWorker struct {
	version
}

func (w *versionWorker) ID() string { return versionWorkerID }

func (w *versionWorker) Stop() error { return nil }

func (w *versionWorker) Start(_ context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	go func() {
		defer close(sensorCh)
		sensorCh <- w
	}()

	return sensorCh, nil
}

func (w *versionWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{&w.version}, nil
}

func newVersionWorker(value string) *versionWorker {
	return &versionWorker{version: version(value)}
}
