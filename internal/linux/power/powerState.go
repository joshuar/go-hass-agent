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
	"strings"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"

	powerStateWorkerID      = "power_state_sensor"
	powerStateWorkerDesc    = "Current power state"
	powerStatePreferencesID = sensorsPrefPrefix + "state"
)

var _ workers.EntityWorker = (*stateWorker)(nil)

var (
	ErrNewPowerStateSensor  = errors.New("could not create power state sensor")
	ErrInitPowerStateWorker = errors.New("could not init power state worker")
)

type powerSignal int

func newPowerState(ctx context.Context, name powerSignal, value any) models.Entity {
	return sensor.NewSensor(ctx,
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
	bus   *dbusx.Bus
	prefs *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

func (w *stateWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitPowerStateWorker,
			fmt.Errorf("unable to set-up D-Bus watch for power state: %w", err))
	}
	sensorCh := make(chan models.Entity)

	// Watch for state changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				switch {
				case strings.HasSuffix(event.Signal, sleepSignal):
					sensorCh <- newPowerState(ctx, suspend, event.Content[0])
				case strings.HasSuffix(event.Signal, shutdownSignal):
					sensorCh <- newPowerState(ctx, shutdown, event.Content[0])
				default:
					continue
				}
			}
		}
	}()

	// Send an initial state update (on, not suspended).
	go func() {
		sensorCh <- newPowerState(ctx, shutdown, false)
	}()

	return sensorCh, nil
}

func (w *stateWorker) PreferencesID() string {
	return powerStatePreferencesID
}

func (w *stateWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *stateWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func NewStateWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitPowerStateWorker, linux.ErrNoSystemBus)
	}

	worker := &stateWorker{
		WorkerMetadata: models.SetWorkerMetadata(powerStateWorkerID, powerStateWorkerDesc),
		bus:            bus,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitPowerStateWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
