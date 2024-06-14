// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver,unexported-return
package device

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

func (v *version) Category() string { return "diagnostic" }

func (v *version) Attributes() any { return nil }

type versionWorker struct{}

func (w *versionWorker) Name() string { return "Go Hass Agent Version Sensor" }

func (w *versionWorker) Description() string {
	return "Sensor displays the current Go Hass Agent version."
}

func (w *versionWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{new(version)}, nil
}

func (w *versionWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	sensors, err := w.Sensors(ctx)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("unable to retrieve version info: %w", err)
	}

	go func() {
		defer close(sensorCh)
		sensorCh <- sensors[0]
	}()

	return sensorCh, nil
}

func newVersionWorker() *versionWorker {
	return &versionWorker{}
}
