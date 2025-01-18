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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"

	powerStateWorkerID      = "power_state_sensor"
	powerStatePreferencesID = "power_state"
)

type powerSignal int

func newPowerState(name powerSignal, value any) sensor.Entity {
	return sensor.NewSensor(
		sensor.WithName("Power State"),
		sensor.WithID("power_state"),
		sensor.WithDeviceClass(types.SensorDeviceClassEnum),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(powerStateIcon(value)),
			sensor.WithValue(powerStateString(name, value)),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			sensor.WithAttribute("options", []string{"Powered On", "Powered Off", "Suspended"}),
		),
		sensor.WithRequestRetry(true),
	)
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
	prefs     *preferences.CommonWorkerPrefs
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

func (w *stateWorker) PreferencesID() string {
	return basePreferencesID + "." + powerStatePreferencesID
}

func (w *stateWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewStateWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(powerStateWorkerID)
	stateWorker := &stateWorker{}

	stateWorker.prefs, err = preferences.LoadWorker(ctx, stateWorker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	if stateWorker.prefs.IsDisabled() {
		return worker, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	stateWorker.triggerCh, err = dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("unable to set-up D-Bus watch for power state: %w", err)
	}

	worker.EventSensorType = stateWorker

	return worker, nil
}
