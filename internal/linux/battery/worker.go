// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var _ workers.EntityWorker = (*BatteryWorker)(nil)

var ErrInitBatterWorker = errors.New("could not init battery worker")

const (
	workerID   = "battery_sensors"
	workerDesc = "Battery statistics"
)

type BatteryWorker struct {
	bus         *dbusx.Bus
	batteryList map[dbus.ObjectPath]context.CancelFunc
	logger      *slog.Logger
	mu          sync.Mutex
	prefs       *WorkerPrefs
	*models.WorkerMetadata
}

func (w *BatteryWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	var wg sync.WaitGroup

	// Get a list of all current connected batteries and monitor them.
	batteries, err := w.getBatteries()
	if err != nil {
		w.logger.Warn("Could not retrieve any battery details from D-Bus.", slog.Any("error", err))
	}

	// For all batteries, start monitoring.
	for _, path := range batteries {
		wg.Add(1)

		go func(path dbus.ObjectPath) {
			defer wg.Done()

			for batterySensor := range w.track(ctx, path) {
				sensorCh <- batterySensor
			}
		}(path)
	}

	wg.Add(1)

	// Send all sensor updates from all tracked batteries to Home Assistant.
	go func() {
		defer wg.Done()

		for batterySensor := range w.monitorBatteryChanges(ctx) {
			sensorCh <- batterySensor
		}
	}()

	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()

	return sensorCh, nil
}

func (w *BatteryWorker) PreferencesID() string {
	return preferencesID
}

func (w *BatteryWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{}
}

func (w *BatteryWorker) IsDisabled() bool {
	return w.prefs.Disabled
}

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func (w *BatteryWorker) getBatteries() ([]dbus.ObjectPath, error) {
	batteryList, err := dbusx.GetData[[]dbus.ObjectPath](w.bus, upowerDBusPath, upowerDBusDest, upowerGetDevicesMethod)
	if err != nil {
		return nil, err
	}

	return batteryList, nil
}

func (w *BatteryWorker) track(ctx context.Context, batteryPath dbus.ObjectPath) <-chan models.Entity {
	w.mu.Lock()
	defer w.mu.Unlock()

	sensorCh := make(chan models.Entity)

	// Ignore if the battery is already being tracked.
	if _, found := w.batteryList[batteryPath]; found {
		slog.Debug("Battery already monitored", slog.String("path", string(batteryPath)))
		close(sensorCh)

		return sensorCh
	}

	var wg sync.WaitGroup

	battery, err := newBattery(w.bus, w.logger, batteryPath)
	if err != nil {
		w.logger.Warn("Cannot monitor battery.",
			slog.Any("path", batteryPath),
			slog.Any("error", err))

		return sensorCh
	}

	battCtx, cancelFunc := context.WithCancel(ctx)

	w.batteryList[batteryPath] = cancelFunc

	wg.Add(1)

	// Get a list of sensors for this battery and send their initial state.
	go func() {
		defer wg.Done()

		for prop := range battery.getSensors(ctx, battery.sensors...) {
			sensorCh <- prop
		}
	}()

	wg.Add(1)

	// Set up a goroutine to monitor for subsequent battery sensor changes.
	go func() {
		defer wg.Done()

		for battery := range monitorBattery(battCtx, battery) {
			sensorCh <- battery
		}
	}()

	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()

	return sensorCh
}

func (w *BatteryWorker) remove(batteryPath dbus.ObjectPath) {
	if cancelFunc, ok := w.batteryList[batteryPath]; ok {
		cancelFunc()
		w.mu.Lock()
		delete(w.batteryList, batteryPath)
		w.mu.Unlock()
	}
}

// monitorBatteryChanges monitors for battery devices being added/removed from
// the system and will start/stop monitory each battery as appropriate.
func (w *BatteryWorker) monitorBatteryChanges(ctx context.Context) <-chan models.Entity {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(upowerDBusPath),
		dbusx.MatchInterface(upowerDBusDest),
		dbusx.MatchMembers(deviceAddedSignal, deviceRemovedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		w.logger.Debug("Unable to set-up D-Bus watch for battery changes.", slog.Any("error", err))

		return nil
	}

	sensorCh := make(chan models.Entity)

	go func() {
		w.logger.Debug("Monitoring for battery additions/removals.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				w.logger.Debug("Stopped monitoring for batteries.")

				return
			case event := <-triggerCh:
				batteryPath, validBatteryPath := event.Content[0].(dbus.ObjectPath)
				if !validBatteryPath {
					continue
				}

				switch {
				case strings.Contains(event.Signal, deviceAddedSignal):
					go func() {
						for s := range w.track(ctx, batteryPath) {
							sensorCh <- s
						}
					}()
				case strings.Contains(event.Signal, deviceRemovedSignal):
					w.remove(batteryPath)
				}
			}
		}
	}()

	return sensorCh
}

func NewBatteryWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	worker := &BatteryWorker{
		WorkerMetadata: models.SetWorkerMetadata(workerID, workerDesc),
		batteryList:    make(map[dbus.ObjectPath]context.CancelFunc),
		bus:            bus,
		logger:         logging.FromContext(ctx).With(slog.String("worker", workerID)),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitBatterWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
