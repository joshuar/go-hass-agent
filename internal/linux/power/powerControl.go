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

//nolint:exhaustruct
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
	entities := make([]*mqtthass.ButtonEntity, 0, len(commands))
	device := linux.MQTTDevice()

	sessionPath, err := dbusx.GetSessionPath(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot create entity.")

		return nil
	}

	for cmdName, cmdInfo := range commands {
		var callback func(p *paho.Publish)
		if cmdInfo.path == "" {
			callback = func(_ *paho.Publish) {
				err := dbusx.Call(ctx, dbusx.SystemBus, sessionPath, loginBaseInterface, cmdInfo.method)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not %s session.", cmdInfo.name)
				}
			}
		} else {
			callback = func(_ *paho.Publish) {
				err := dbusx.Call(ctx, dbusx.SystemBus, cmdInfo.path, loginBaseInterface, cmdInfo.method, true)
				if err != nil {
					log.Warn().Err(err).Msg("Could not power off session.")
				}
			}
		}

		entities = append(entities,
			mqtthass.AsButton(
				mqtthass.NewEntity(preferences.AppName, cmdInfo.name, device.Name+"_"+cmdName).
					WithOriginInfo(preferences.MQTTOrigin()).
					WithDeviceInfo(device).
					WithIcon(cmdInfo.icon).
					WithCommandCallback(callback)))
	}

	return entities
}
