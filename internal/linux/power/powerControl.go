// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusSessionLockMethod      = sessionInterface + ".Lock"
	dbusSessionUnlockMethod    = sessionInterface + ".UnLock"
	dbusSessionRebootMethod    = managerInterface + ".Reboot"
	dbusSessionSuspendMethod   = managerInterface + ".Suspend"
	dbusSessionHibernateMethod = managerInterface + ".Hibernate"
	dbusSessionPowerOffMethod  = managerInterface + ".PowerOff"
)

type commandConfig struct {
	name   string
	icon   string
	path   string
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
		path:   loginBasePath,
		method: dbusSessionRebootMethod,
	},
	"suspend": {
		name:   "Suspend",
		icon:   "mdi:power-sleep",
		path:   loginBasePath,
		method: dbusSessionSuspendMethod,
	},
	"hibernate": {
		name:   "Hibernate",
		icon:   "mdi:power-sleep",
		path:   loginBasePath,
		method: dbusSessionHibernateMethod,
	},
	"power_off": {
		name:   "Power Off",
		icon:   "mdi:power",
		path:   loginBasePath,
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
				err := dbusx.Call(ctx, dbusx.SystemBus, string(sessionPath), loginBaseInterface, v.method)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not %s session.", v.name)
				}
			}
		} else {
			callback = func(_ *paho.Publish) {
				err := dbusx.Call(ctx, dbusx.SystemBus, v.path, loginBaseInterface, v.method, true)
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
