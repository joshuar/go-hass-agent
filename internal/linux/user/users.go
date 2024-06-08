// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

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
)

type usersSensor struct {
	userNames []string
	linux.Sensor
}

func (s *usersSensor) Attributes() any {
	return struct {
		DataSource string   `json:"data_source"`
		Usernames  []string `json:"usernames"`
	}{
		DataSource: linux.DataSrcDbus,
		Usernames:  s.userNames,
	}
}

func (s *usersSensor) updateUsers(ctx context.Context) error {
	userData, err := dbusx.GetData[[][]any](ctx, dbusx.SystemBus, loginBasePath, loginBaseInterface, listSessionsMethod)
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
func newUsersSensor() *usersSensor {
	userSensor := &usersSensor{}
	userSensor.SensorTypeValue = linux.SensorUsers
	userSensor.UnitsString = "users"
	userSensor.IconString = "mdi:account"
	userSensor.StateClassValue = types.StateClassMeasurement

	return userSensor
}

type worker struct {
	sensor *usersSensor
}

//nolint:exhaustruct
func (w *worker) Setup(_ context.Context) *dbusx.Watch {
	return &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionAddedSignal, sessionRemovedSignal},
		Interface: managerInterface,
		Path:      loginBasePath,
	}
}

func (w *worker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	sendUpdate := func() {
		err := w.sensor.updateUsers(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Update failed.")

			return
		}
		sensorCh <- w.sensor
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped users sensors.")

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

	return sensorCh
}

func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	err := w.sensor.updateUsers(ctx)

	return []sensor.Details{w.sensor}, err
}

func NewUserWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "User count sensor",
			WorkerDesc: "Sensors for number of logged in users.",
			Value: &worker{
				sensor: newUsersSensor(),
			},
		},
		nil
}
