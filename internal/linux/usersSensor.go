// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

const (
	login1DBusPath = "/org/freedesktop/login1"
)

type usersSensor struct {
	userNames []string
	linuxSensor
}

func (s *usersSensor) Attributes() interface{} {
	return struct {
		DataSource string   `json:"Data Source"`
		Usernames  []string `json:"Usernames"`
	}{
		DataSource: srcDbus,
		Usernames:  s.userNames,
	}
}

func UsersUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	updateUsers := func() {
		sensor := newUsersSensor()
		userData := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
			Path(login1DBusPath).
			Destination("org.freedesktop.login1").
			GetData("org.freedesktop.login1.Manager.ListUsers").AsRawInterface()
		if userList, ok := userData.([][]interface{}); !ok {
			return
		} else {
			sensor.value = len(userList)
			for _, u := range userList {
				sensor.userNames = append(sensor.userNames, u[1].(string))
			}
			sensorCh <- sensor
		}
	}
	updateUsers()

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(login1DBusPath),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case "org.freedesktop.login1.Manager.SessionNew",
				"org.freedesktop.login1.Manager.SessionRemoved":
				updateUsers()
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
	}()
	return sensorCh
}

func newUsersSensor() *usersSensor {
	s := &usersSensor{}
	s.sensorType = users
	s.units = "users"
	s.icon = "mdi:account"
	s.stateClass = sensor.StateMeasurement
	return s
}
