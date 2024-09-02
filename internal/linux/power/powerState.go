// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

func newPowerState(signalName powerSignal, signalValue any) *powerStateSensor {
	return &powerStateSensor{
		signal: signalName,
		Sensor: linux.Sensor{
			DisplayName:  "Power State",
			Value:        signalValue,
			DataSource:   linux.DataSrcDbus,
			IsDiagnostic: true,
		},
	}
}

type stateWorker struct {
	logger    *slog.Logger
	triggerCh chan dbusx.Trigger
}

func (w *stateWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	// Watch for state changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
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

func NewStateWorker(ctx context.Context) (*linux.SensorWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to set-up D-Bus watch for power state: %w", err)
	}

	return &linux.SensorWorker{
			Value: &stateWorker{
				triggerCh: triggerCh,
			},
			WorkerID: powerStateWorkerID,
		},
		nil
}
