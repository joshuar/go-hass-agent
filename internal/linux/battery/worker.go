// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package battery

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type sensorWorker struct {
	bus         *dbusx.Bus
	batteryList map[dbus.ObjectPath]context.CancelFunc
	logger      *slog.Logger
	mu          sync.Mutex
}

// ?: implement initial battery sensor retrieval.
//
//revive:disable:unused-receiver
func (w *sensorWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return nil, linux.ErrUnimplemented
}

func (w *sensorWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	var wg sync.WaitGroup

	// Get a list of all current connected batteries and monitor them.
	batteries, err := w.getBatteries()
	if err != nil {
		w.logger.Warn("Could not retrieve any battery details from D-Bus.", slog.Any("error", err))
	}

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

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func (w *sensorWorker) getBatteries() ([]dbus.ObjectPath, error) {
	batteryList, err := dbusx.GetData[[]dbus.ObjectPath](w.bus, upowerDBusPath, upowerDBusDest, upowerGetDevicesMethod)
	if err != nil {
		return nil, err
	}

	return batteryList, nil
}

func (w *sensorWorker) track(ctx context.Context, batteryPath dbus.ObjectPath) <-chan sensor.Entity {
	sensorCh := make(chan sensor.Entity)

	var wg sync.WaitGroup

	battery, err := newBattery(w.bus, w.logger, batteryPath)
	if err != nil {
		w.logger.Warn("Cannot monitor battery.",
			slog.Any("path", batteryPath),
			slog.Any("error", err))

		return sensorCh
	}

	battCtx, cancelFunc := context.WithCancel(ctx)

	w.mu.Lock()
	w.batteryList[batteryPath] = cancelFunc
	w.mu.Unlock()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for prop := range battery.getSensors(battery.sensors...) {
			sensorCh <- prop
		}
	}()

	wg.Add(1)

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

func (w *sensorWorker) remove(batteryPath dbus.ObjectPath) {
	if cancelFunc, ok := w.batteryList[batteryPath]; ok {
		cancelFunc()
		w.mu.Lock()
		delete(w.batteryList, batteryPath)
		w.mu.Unlock()
	}
}

// monitorBatteryChanges monitors for battery devices being added/removed from
// the system and will start/stop monitory each battery as appropriate.
func (w *sensorWorker) monitorBatteryChanges(ctx context.Context) <-chan sensor.Entity {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(upowerDBusPath),
		dbusx.MatchInterface(upowerDBusDest),
		dbusx.MatchMembers(deviceAddedSignal, deviceRemovedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		w.logger.Debug("Unable to set-up D-Bus watch for battery changes.", slog.Any("error", err))

		return nil
	}

	sensorCh := make(chan sensor.Entity)

	// events, err := dbusx.NewWatch(
	// 	dbusx.MatchPath(upowerDBusPath),
	// 	dbusx.MatchInterface(upowerDBusDest),
	// 	dbusx.MatchMember(deviceAddedSignal, deviceRemovedSignal),
	// ).Start(ctx, w.bus)
	// if err != nil {
	// 	w.logger.Debug("Failed to create D-Bus watch for battery additions/removals.", "error", err.Error())
	// 	close(sensorCh)

	// 	return sensorCh
	// }

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

func NewBatteryWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(workerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	worker.EventType = &sensorWorker{
		batteryList: make(map[dbus.ObjectPath]context.CancelFunc),
		bus:         bus,
		logger:      logging.FromContext(ctx).With(slog.String("worker", workerID)),
	}

	return worker, nil
}
