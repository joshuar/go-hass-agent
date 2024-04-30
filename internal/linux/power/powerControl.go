// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/godbus/dbus/v5"
	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusSessionDest            = "org.freedesktop.login1"
	dbusSessionLockMethod      = dbusSessionDest + ".Session.Lock"
	dbusSessionUnlockMethod    = dbusSessionDest + ".Session.UnLock"
	dbusSessionRebootMethod    = dbusSessionDest + ".Manager.Reboot"
	dbusSessionSuspendMethod   = dbusSessionDest + ".Manager.Suspend"
	dbusSessionHibernateMethod = dbusSessionDest + ".Manager.Hibernate"
	dbusSessionPowerOffMethod  = dbusSessionDest + ".Manager.PowerOff"
)

type commandConfig struct {
	name   string
	icon   string
	path   dbus.ObjectPath
	method string
}

var commands = map[string]commandConfig{
	"lock_session": {
		name:   "lock",
		icon:   "mdi:eye-lock",
		method: dbusSessionLockMethod,
	},
	"unlock_session": {
		name:   "unlock",
		icon:   "mdi:eye-lock-open",
		method: dbusSessionUnlockMethod,
	},
	"reboot": {
		name:   "reboot",
		icon:   "mdi:restart",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionRebootMethod,
	},
	"suspend": {
		name:   "suspend",
		icon:   "mdi:power-sleep",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionSuspendMethod,
	},
	"hibernate": {
		name:   "hibernate",
		icon:   "mdi:power-sleep",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionHibernateMethod,
	},
	"poweroff": {
		name:   "power off",
		icon:   "mdi:power",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionPowerOffMethod,
	},
}

func NewPowerControl(ctx context.Context) []*mqtthass.EntityConfig {
	var entities []*mqtthass.EntityConfig
	sessionPath := dbusx.GetSessionPath(ctx)

	for k, v := range commands {
		var callback func(MQTT.Client, MQTT.Message)
		if v.path == "" {
			callback = func(_ MQTT.Client, _ MQTT.Message) {
				err := systemDBusCall(ctx, sessionPath, dbusSessionDest, v.method)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not %s session.", v.name)
				}
			}
		} else {
			callback = func(_ MQTT.Client, _ MQTT.Message) {
				err := systemDBusCall(ctx, v.path, dbusSessionDest, v.method, true)
				if err != nil {
					log.Warn().Err(err).Msg("Could not power off session.")
				}
			}
		}
		entities = append(entities, linux.NewButton(k).
			WithIcon(v.icon).
			WithCommandCallback(callback))
	}
	return entities
}

func systemDBusCall(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
	return dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(path).
		Destination(dest).
		Call(method, args...)
}
