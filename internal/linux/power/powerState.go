// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"

	powerStateWorkerID      = "power_state_sensor"
	powerStatePreferencesID = sensorsPrefPrefix + "state"
)

var (
	ErrNewPowerStateSensor  = errors.New("could not create power state sensor")
	ErrInitPowerStateWorker = errors.New("could not init power state worker")
)

type powerSignal int

func newPowerState(ctx context.Context, name powerSignal, value any) (*models.Entity, error) {
	stateSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Power State"),
		sensor.WithID("power_state"),
		sensor.WithDeviceClass(class.SensorClassEnum),
		sensor.AsDiagnostic(),
		sensor.WithIcon(powerStateIcon(value)),
		sensor.WithState(powerStateString(name, value)),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		sensor.WithAttribute("options", []string{"Powered On", "Powered Off", "Suspended"}),
		sensor.AsRetryableRequest(true),
	)
	if err != nil {
		return nil, errors.Join(ErrNewPowerStateSensor, err)
	}

	return &stateSensor, nil
}

func powerStateString(signal powerSignal, value any) string {
	state, ok := value.(bool)
	if !ok {
		return "Unknown"
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

func (w *stateWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	// Watch for state changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				var (
					entity *models.Entity
					err    error
				)

				switch {
				case strings.HasSuffix(event.Signal, sleepSignal):
					entity, err = newPowerState(ctx, suspend, event.Content[0])
				case strings.HasSuffix(event.Signal, shutdownSignal):
					entity, err = newPowerState(ctx, shutdown, event.Content[0])
				}

				if err != nil || entity == nil {
					logging.FromContext(ctx).Warn("Could not generate power state sensor.",
						slog.Any("error", err))
					continue
				}

				sensorCh <- *entity
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
func (w *stateWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	entity, err := newPowerState(ctx, shutdown, false)
	if err != nil {
		return nil, fmt.Errorf("could not generate power state sensor: %w", err)
	}

	return []models.Entity{*entity}, nil
}

func (w *stateWorker) PreferencesID() string {
	return powerStatePreferencesID
}

func (w *stateWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewStateWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(powerStateWorkerID)
	stateWorker := &stateWorker{}

	stateWorker.prefs, err = preferences.LoadWorker(stateWorker)
	if err != nil {
		return nil, errors.Join(ErrInitPowerStateWorker, err)
	}

	if stateWorker.prefs.IsDisabled() {
		return worker, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, errors.Join(ErrInitPowerStateWorker, linux.ErrNoSystemBus)
	}

	stateWorker.triggerCh, err = dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, bus)
	if err != nil {
		return worker, errors.Join(ErrInitPowerStateWorker,
			fmt.Errorf("unable to set-up D-Bus watch for power state: %w", err))
	}

	worker.EventSensorType = stateWorker

	return worker, nil
}
