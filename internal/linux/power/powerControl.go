// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"

	"github.com/eclipse/paho.golang/paho"
	"github.com/godbus/dbus/v5"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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
		name:   "Lock Session",
		icon:   "mdi:eye-lock",
		method: dbusSessionLockMethod,
	},
	"unlock_session": {
		name:   "Unlock Session",
		icon:   "mdi:eye-lock-open",
		method: dbusSessionUnlockMethod,
	},
	"reboot": {
		name:   "Reboot",
		icon:   "mdi:restart",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionRebootMethod,
	},
	"suspend": {
		name:   "Suspend",
		icon:   "mdi:power-sleep",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionSuspendMethod,
	},
	"hibernate": {
		name:   "Hibernate",
		icon:   "mdi:power-sleep",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionHibernateMethod,
	},
	"power_off": {
		name:   "Power Off",
		icon:   "mdi:power",
		path:   dbus.ObjectPath("/org/freedesktop/login1"),
		method: dbusSessionPowerOffMethod,
	},
}

func NewPowerControl(ctx context.Context) []*mqtthass.ButtonEntity {
	var entities []*mqtthass.ButtonEntity
	device := linux.MQTTDevice()
	sessionPath := dbusx.GetSessionPath(ctx)

	for k, v := range commands {
		var callback func(p *paho.Publish)
		if v.path == "" {
			callback = func(_ *paho.Publish) {
				err := systemDBusCall(ctx, sessionPath, dbusSessionDest, v.method)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not %s session.", v.name)
				}
			}
		} else {
			callback = func(_ *paho.Publish) {
				err := systemDBusCall(ctx, v.path, dbusSessionDest, v.method, true)
				if err != nil {
					log.Warn().Err(err).Msg("Could not power off session.")
				}
			}
		}
		entities = append(entities,
			mqtthass.AsButton(
				mqtthass.NewEntity(preferences.AppName, v.name, device.Name+"_"+k).
					WithOriginInfo(preferences.MQTTOrigin()).
					WithDeviceInfo(device).
					WithIcon(v.icon).
					WithCommandCallback(callback)))
	}
	return entities
}

func systemDBusCall(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
	return dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(path).
		Destination(dest).
		Call(method, args...)
}
