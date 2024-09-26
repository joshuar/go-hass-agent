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
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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

func newPowerState(name powerSignal, value any) sensor.Entity {
	return sensor.Entity{
		Name:     "Power State",
		Category: types.CategoryDiagnostic,
		EntityState: &sensor.EntityState{
			ID:    "power_state",
			Icon:  powerStateIcon(value),
			State: powerStateString(name, value),
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
			},
		},
	}
}

func powerStateString(signal powerSignal, value any) string {
	state, ok := value.(bool)
	if !ok {
		return sensor.StateUnknown
	}

	switch {
	case signal == suspend && state:
		return "Suspended"
	case signal == shutdown && state:
		return "Powered Off"
	default:
		return "Powered On"
	}
}

func powerStateIcon(value any) string {
	state, ok := value.(string)
	if !ok {
		return "mdi:power-on"
	}

	switch state {
	case "Suspended":
		return "mdi:power-sleep"
	case "Powered Off":
		return "mdi:power-off"
	default:
		return "mdi:power-on"
	}
}

type stateWorker struct {
	triggerCh chan dbusx.Trigger
}

func (w *stateWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	// Watch for state changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				switch {
				case strings.HasSuffix(event.Signal, sleepSignal):
					sensorCh <- newPowerState(suspend, event.Content[0])
				case strings.HasSuffix(event.Signal, shutdownSignal):
					sensorCh <- newPowerState(shutdown, event.Content[0])
				}
			}
		}
	}()

	// Send an initial state update (on, not suspended).
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", powerStateWorkerID)).
				Debug("Could not retrieve power state.", slog.Any("error", err))

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
func (w *stateWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return []sensor.Entity{newPowerState(shutdown, false)}, nil
}

func NewStateWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(powerStateWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("unable to set-up D-Bus watch for power state: %w", err)
	}

	worker.EventType = &stateWorker{
		triggerCh: triggerCh,
	}

	return worker, nil
}
