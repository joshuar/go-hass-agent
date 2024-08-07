// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package user

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
	loginBasePath        = "/org/freedesktop/login1"
	loginBaseInterface   = "org.freedesktop.login1"
	managerInterface     = loginBaseInterface + ".Manager"
	sessionAddedSignal   = "SessionNew"
	sessionRemovedSignal = "SessionRemoved"
	listSessionsMethod   = managerInterface + ".ListSessions"

	usersWorkerID = "users_sensors"
)

type usersSensor struct {
	bus       *dbusx.Bus
	userNames []string
	linux.Sensor
}

func (s *usersSensor) Attributes() map[string]any {
	attributes := make(map[string]any)
	attributes["data_source"] = linux.DataSrcDbus
	attributes["usernames"] = s.userNames

	return attributes
}

func (s *usersSensor) updateUsers(ctx context.Context) error {
	userData, err := dbusx.GetData[[][]any](ctx, s.bus, loginBasePath, loginBaseInterface, listSessionsMethod)
	if err != nil {
		return fmt.Errorf("could not retrieve users from D-Bus: %w", err)
	}

	s.Value = len(userData)

	var users []string

	for _, u := range userData {
		if user, ok := u[2].(string); ok {
			users = append(users, user)
		}
	}

	s.userNames = users

	return nil
}

//nolint:exhaustruct
func newUsersSensor(bus *dbusx.Bus) *usersSensor {
	userSensor := &usersSensor{bus: bus}
	userSensor.SensorTypeValue = linux.SensorUsers
	userSensor.UnitsString = "users"
	userSensor.IconString = "mdi:account"
	userSensor.StateClassValue = types.StateClassMeasurement

	return userSensor
}

type worker struct {
	sensor *usersSensor
	logger *slog.Logger
	bus    *dbusx.Bus
}

//nolint:exhaustruct
func (w *worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	triggerCh, err := w.bus.WatchBus(ctx, &dbusx.Watch{
		Names:     []string{sessionAddedSignal, sessionRemovedSignal},
		Interface: managerInterface,
		Path:      loginBasePath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for user updates: %w", err)
	}

	sendUpdate := func() {
		err := w.sensor.updateUsers(ctx)
		if err != nil {
			w.logger.Debug("Update failed", "error", err.Error())

			return
		}
		sensorCh <- w.sensor
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				if !strings.Contains(event.Signal, sessionAddedSignal) && !strings.Contains(event.Signal, sessionRemovedSignal) {
					continue
				}

				go sendUpdate()
			}
		}
	}()

	// Send an initial sensor update.
	go sendUpdate()

	return sensorCh, nil
}

func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	err := w.sensor.updateUsers(ctx)

	return []sensor.Details{w.sensor}, err
}

func NewUserWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor power profile: %w", err)
	}

	return &linux.SensorWorker{
			Value: &worker{
				sensor: newUsersSensor(bus),
				logger: logging.FromContext(ctx).With(slog.String("worker", usersWorkerID)),
				bus:    bus,
			},
			WorkerID: usersWorkerID,
		},
		nil
}
