// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package user

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

const (
	login1DBusPath = "/org/freedesktop/login1"
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
	userData := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(login1DBusPath).
		Destination("org.freedesktop.login1").
		GetData("org.freedesktop.login1.Manager.ListUsers").AsRawInterface()
	var userList [][]any
	var ok bool
	if userList, ok = userData.([][]any); !ok {
		return
	}
	s.Value = len(userList)
	var users []string
	for _, u := range userList {
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
	s.StateClassValue = sensor.StateMeasurement
	return s
}

func Updater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	u := newUsersSensor()
	u.updateUsers(ctx)
	sensorCh <- u

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(login1DBusPath),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case "org.freedesktop.login1.Manager.SessionNew",
				"org.freedesktop.login1.Manager.SessionRemoved":
				u.updateUsers(ctx)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Failed to create user D-Bus watch. Users sensor will not run.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped users sensors.")
	}()
	return sensorCh
}
