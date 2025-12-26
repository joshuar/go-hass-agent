// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package power

import (
	"context"
	"fmt"
	"strings"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	suspend powerSignal = iota
	shutdown
)

const (
	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"

	powerStateWorkerID      = "power_state_sensor"
	powerStateWorkerDesc    = "Current power state"
	powerStatePreferencesID = sensorsPrefPrefix + "state"
)

var _ workers.EntityWorker = (*stateWorker)(nil)

type powerSignal int

func newPowerState(ctx context.Context, name powerSignal, value any) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Power State"),
		sensor.WithID("power_state"),
		sensor.WithDeviceClass(class.SensorClassEnum),
		sensor.AsDiagnostic(),
		sensor.WithIcon(powerStateIcon(value)),
		sensor.WithState(powerStateString(name, value)),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
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
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *workers.CommonWorkerPrefs
}

func NewStateWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &stateWorker{
		WorkerMetadata: models.SetWorkerMetadata(powerStateWorkerID, powerStateWorkerDesc),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(powerStatePreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *stateWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch power state: %w", err)
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

func (w *stateWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
