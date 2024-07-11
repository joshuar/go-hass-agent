// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"

	powerStateWorkerID = "power_state_sensor"
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

type stateWorker struct {
	logger *slog.Logger
	bus    *dbusx.Bus
}

//nolint:exhaustruct
func (w *stateWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	triggerCh, err := w.bus.WatchBus(ctx, &dbusx.Watch{
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
			w.logger.Debug("Could not retrieve power state.", "error", err.Error())

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

func NewStateWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor power state: %w", err)
	}

	return &linux.SensorWorker{
			Value: &stateWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", powerStateWorkerID)),
				bus:    bus,
			},
			WorkerID: powerStateWorkerID,
		},
		nil
}
