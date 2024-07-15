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

func (v *version) Attributes() map[string]any { return nil }

type VersionWorker struct{}

func (w *VersionWorker) ID() string { return versionWorkerID }

// Stop is a no-op.
func (w *VersionWorker) Stop() error { return nil }

func (w *VersionWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{new(version)}, nil
}

func (w *VersionWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
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

func NewVersionWorker() *VersionWorker {
	return &VersionWorker{}
}
