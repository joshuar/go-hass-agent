// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
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

func UsersUpdater(ctx context.Context, tracker device.SensorTracker) {
	updateUsers := func() {
		sensor := newUsersSensor()
		userData := NewBusRequest(ctx, SystemBus).
			Path(login1DBusPath).
			Destination("org.freedesktop.login1").
			GetData("org.freedesktop.login1.Manager.ListUsers").AsRawInterface()
		userList := userData.([][]interface{})
		sensor.value = len(userList)
		for _, u := range userList {
			sensor.userNames = append(sensor.userNames, u[1].(string))
		}
		if err := tracker.UpdateSensors(ctx, sensor); err != nil {
			log.Error().Err(err).Msg("Could not update users sensor.")
		}
	}
	updateUsers()

	err := NewBusRequest(ctx, SystemBus).
		Path(login1DBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(login1DBusPath),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case "org.freedesktop.login1.Manager.SessionNew":
			case "org.freedesktop.login1.Manager.SessionRemoved":
				updateUsers()
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create user D-Bus watch.")
	}
}

func newUsersSensor() *usersSensor {
	s := &usersSensor{}
	s.sensorType = users
	s.units = "users"
	s.icon = "mdi:account"
	s.stateClass = sensor.StateMeasurement
	return s
}
