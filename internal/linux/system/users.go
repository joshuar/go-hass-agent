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

func newUsersSensor(users []string) sensor.Entity {
	return sensor.Entity{
		Name:       "Current Users",
		StateClass: types.StateClassMeasurement,
		Units:      sensorUnits,
		State: &sensor.State{
			ID:    "current_users",
			Icon:  sensorIcon,
			Value: len(users),
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
				"usernames":   users,
			},
		},
	}
}

type Worker struct {
	getUsers  func() ([]string, error)
	triggerCh chan dbusx.Trigger
	linux.EventSensorWorker
}

func (w *Worker) Events(ctx context.Context) (chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	sendUpdate := func() {
		users, err := w.getUsers()
		if err != nil {
			slog.With(slog.String("worker", usersWorkerID)).Debug("Failed to get list of user sessions.", slog.Any("error", err))
		} else {
			sensorCh <- newUsersSensor(users)
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

func (w *Worker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	users, err := w.getUsers()

	return []sensor.Entity{newUsersSensor(users)}, err
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

	worker.getUsers = func() ([]string, error) {
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

	return worker, nil
}
