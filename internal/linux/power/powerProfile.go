// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	powerProfilesPath      = "/net/hadess/PowerProfiles"
	powerProfilesDest      = "net.hadess.PowerProfiles"
	powerProfilesInterface = "org.freedesktop.Upower.PowerProfiles"
	activeProfileProp      = "ActiveProfile"
)

type powerSensor struct {
	linux.Sensor
}

//nolint:exhaustruct
func newPowerSensor(sensorType linux.SensorTypeValue, sensorValue dbus.Variant) *powerSensor {
	newSensor := &powerSensor{}
	newSensor.Value = dbusx.VariantToValue[string](sensorValue)
	newSensor.SensorTypeValue = sensorType
	newSensor.IconString = "mdi:flash"
	newSensor.SensorSrc = linux.DataSrcDbus
	newSensor.IsDiagnostic = true

	return newSensor
}

type profileWorker struct{}

//nolint:exhaustruct
func (w *profileWorker) Setup(_ context.Context) (*dbusx.Watch, error) {
	return &dbusx.Watch{
			Bus:       dbusx.SystemBus,
			Names:     []string{dbusx.PropChangedSignal},
			Interface: dbusx.PropInterface,
			Path:      powerProfilesPath,
		},
		nil
}

func (w *profileWorker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	// Check for power profile support, exit if not available. Otherwise, send
	// an initial update.
	sensors, err := w.Sensors(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot monitor power profile.")
		close(sensorCh)

		return sensorCh
	}

	go func() {
		sensorCh <- sensors[0]
	}()

	// Watch for power profile changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg(("Stopped power profile sensor."))

				return
			case event := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					log.Warn().Err(err).Msg("Did not understand received trigger.")

					continue
				}

				if profile, profileChanged := props.Changed[activeProfileProp]; profileChanged {
					sensorCh <- newPowerSensor(linux.SensorPowerProfile, profile)
				}
			}
		}
	}()

	return sensorCh
}

func (w *profileWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	profile, err := dbusx.GetProp[dbus.Variant](ctx,
		dbusx.SystemBus,
		powerProfilesPath,
		powerProfilesDest,
		powerProfilesDest+"."+activeProfileProp)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve a power profile from D-Bus: %w", err)
	}

	return []sensor.Details{newPowerSensor(linux.SensorPowerProfile, profile)}, nil
}

func NewProfileWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Power Profile Sensor",
			WorkerDesc: "Sensor to track the current power profile.",
			Value:      &profileWorker{},
		},
		nil
}
