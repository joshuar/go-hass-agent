// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package user

import (
	"context"

	"github.com/godbus/dbus/v5"
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
	sessionAddedSignal   = managerInterface + ".SessionNew"
	sessionRemovedSignal = managerInterface + ".SessionRemoved"
	listUsersMethod      = managerInterface + ".ListUsers"
)

type usersSensor struct {
	userNames []string
	linux.Sensor
}

func (s *usersSensor) Attributes() any {
	return struct {
		DataSource string   `json:"Data Source"`
		Usernames  []string `json:"Usernames"`
	}{
		DataSource: linux.DataSrcDbus,
		Usernames:  s.userNames,
	}
}

func (s *usersSensor) updateUsers(ctx context.Context) {
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(loginBasePath).
		Destination(loginBaseInterface)

	userData, err := dbusx.GetData[[][]any](req, listUsersMethod)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve users from D-Bus.")
	}
	s.Value = len(userData)
	var users []string
	for _, u := range userData {
		if user, ok := u[1].(string); ok {
			users = append(users, user)
		}
	}
	s.userNames = users
}

func newUsersSensor() *usersSensor {
	s := &usersSensor{}
	s.SensorTypeValue = linux.SensorUsers
	s.UnitsString = "users"
	s.IconString = "mdi:account"
	s.StateClassValue = types.StateClassMeasurement
	return s
}

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	u := newUsersSensor()

	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(loginBasePath),
			dbus.WithMatchInterface(dbusx.PropInterface),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case sessionAddedSignal, sessionRemovedSignal:
				u.updateUsers(ctx)
				go func() {
					sensorCh <- u
				}()
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to monitor for user login/logout.")
		close(sensorCh)
		return sensorCh
	}

	// Send an initial sensor update.
	u.updateUsers(ctx)
	go func() {
		sensorCh <- u
	}()

	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped users sensors.")
	}()
	return sensorCh
}
