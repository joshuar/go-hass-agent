// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	screensaverDBusPath      = "/org/freedesktop/ScreenSaver"
	screensaverDBusInterface = "org.freedesktop.ScreenSaver"
)

type screenlockSensor struct {
	linuxSensor
}

func (s *screenlockSensor) Icon() string {
	if s.value.(bool) {
		return "mdi:eye-lock"
	} else {
		return "mdi:eye-lock-open"
	}
}

func (s *screenlockSensor) SensorType() sensor.SensorType {
	return sensor.TypeBinary
}

func ScreenLockUpdater(ctx context.Context, update chan interface{}) {
	err := NewBusRequest(SessionBus).
		Path(screensaverDBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(screensaverDBusPath),
			dbus.WithMatchInterface(screensaverDBusInterface),
		}).
		Event("org.freedesktop.ScreenSaver.ActiveChanged").
		Handler(func(s *dbus.Signal) {
			lock := &screenlockSensor{}
			lock.value = s.Body[0].(bool)
			lock.sensorType = screenLock
			lock.source = SOURCE_DBUS
			update <- lock
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create screen lock DBus watch.")
	}
}
