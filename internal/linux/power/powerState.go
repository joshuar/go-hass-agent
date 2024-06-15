// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"
)

type powerSignal int

type powerStateSensor struct {
	linux.Sensor
	signal powerSignal
}

func (s *powerStateSensor) State() any {
	boolVal, ok := s.Value.(bool)
	if !ok {
		return sensor.StateUnknown
	}

	if boolVal {
		switch s.signal {
		case suspend:
			return "Suspended"
		case shutdown:
			return "Powered Off"
		}
	}

	return "Powered On"
}

func (s *powerStateSensor) Icon() string {
	str, ok := s.State().(string)
	if !ok {
		str = ""
	}

	switch str {
	case "Suspended":
		return "mdi:power-sleep"
	case "Powered Off":
		return "mdi:power-off"
	default:
		return "mdi:power-on"
	}
}

//nolint:exhaustruct
func newPowerState(signalName powerSignal, signalValue any) *powerStateSensor {
	return &powerStateSensor{
		signal: signalName,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorPowerState,
			Value:           signalValue,
			SensorSrc:       linux.DataSrcDbus,
			IsDiagnostic:    true,
		},
	}
}

type stateWorker struct{}

//nolint:exhaustruct
func (w *stateWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	triggerCh, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sleepSignal, shutdownSignal},
		Interface: managerInterface,
		Path:      loginBasePath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for power state updates: %w", err)
	}

	// Watch for state changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped power state sensor.")

				return
			case event := <-triggerCh:
				switch event.Signal {
				case sleepSignal:
					go func() {
						sensorCh <- newPowerState(suspend, event.Content[0])
					}()
				case shutdownSignal:
					go func() {
						sensorCh <- newPowerState(shutdown, event.Content[0])
					}()
				}
			}
		}
	}()

	// Send an initial state update (on, not suspended).
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to retrieve sensors.")

			return
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

// Sensors returns the current sensors states. Assuming that if this is called,
// then the machine is obviously running and not suspended, otherwise this
// couldn't be called?
func (w *stateWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{newPowerState(shutdown, false)}, nil
}

func NewStateWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Power State Sensor",
			WorkerDesc: "Sensor to track the current power state of the device.",
			Value:      &stateWorker{},
		},
		nil
}
