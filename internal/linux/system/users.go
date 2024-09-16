// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	loginBasePath        = "/org/freedesktop/login1"
	loginBaseInterface   = "org.freedesktop.login1"
	managerInterface     = loginBaseInterface + ".Manager"
	sessionAddedSignal   = "SessionNew"
	sessionRemovedSignal = "SessionRemoved"
	listSessionsMethod   = managerInterface + ".ListSessions"

	sensorUnits = "users"
	sensorIcon  = "mdi:account"

	usersWorkerID = "users_sensors"
)

type usersSensor struct {
	getUsers  func() ([]string, error)
	userNames []string
	linux.Sensor
}

func (s *usersSensor) State() any {
	return len(s.userNames)
}

func (s *usersSensor) Attributes() map[string]any {
	attributes := s.Sensor.Attributes()
	attributes["usernames"] = s.userNames

	return attributes
}

func newUsersSensor() *usersSensor {
	return &usersSensor{
		Sensor: linux.Sensor{
			DisplayName:     "Current Users",
			UniqueID:        "current_users",
			UnitsString:     sensorUnits,
			IconString:      sensorIcon,
			StateClassValue: types.StateClassMeasurement,
			DataSource:      linux.DataSrcDbus,
		},
	}
}

type Worker struct {
	sensor    *usersSensor
	triggerCh chan dbusx.Trigger
	linux.EventSensorWorker
}

func (w *Worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	sendUpdate := func() {
		users, err := w.sensor.getUsers()
		if err != nil {
			slog.With(slog.String("worker", usersWorkerID)).Debug("Failed to get list of user sessions.", slog.Any("error", err))
		} else {
			w.sensor.userNames = users
			sensorCh <- w.sensor
		}
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.triggerCh:
				go sendUpdate()
			}
		}
	}()

	// Send an initial sensor update.
	go sendUpdate()

	return sensorCh, nil
}

func (w *Worker) Sensors(_ context.Context) ([]sensor.Details, error) {
	users, err := w.sensor.getUsers()
	w.sensor.userNames = users

	return []sensor.Details{w.sensor}, err
}

func NewUserWorker(ctx context.Context) (*Worker, error) {
	worker := &Worker{}
	worker.WorkerID = usersWorkerID

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sessionAddedSignal, sessionRemovedSignal),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to set-up D-Bus watch for user sessions: %w", err)
	}

	worker.triggerCh = triggerCh

	usersSensor := newUsersSensor()

	usersSensor.getUsers = func() ([]string, error) {
		userData, err := dbusx.GetData[[][]any](bus, loginBasePath, loginBaseInterface, listSessionsMethod)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve users from D-Bus: %w", err)
		}

		var users []string

		for _, u := range userData {
			if user, ok := u[2].(string); ok {
				users = append(users, user)
			}
		}

		return users, nil
	}

	worker.sensor = usersSensor

	return worker, nil
}
