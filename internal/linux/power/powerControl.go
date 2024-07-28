// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"log/slog"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"

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

func NewPowerControl(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger, device *mqtthass.Device) []*mqtthass.ButtonEntity {
	logger := parentLogger.With(slog.String("controller", "power"))

	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		logger.Warn("Cannot create power controls.", "error", err.Error())

		return nil
	}

	sessionPath, err := bus.GetSessionPath(ctx)
	if err != nil {
		logger.Warn("Cannot create power controls.", "error", err.Error())

		return nil
	}

	entities := make([]*mqtthass.ButtonEntity, 0, len(commands))

	for cmdName, cmdInfo := range commands {
		var callback func(p *paho.Publish)
		if cmdInfo.path == "" {
			callback = func(_ *paho.Publish) {
				err := bus.Call(ctx, sessionPath, loginBaseInterface, cmdInfo.method)
				if err != nil {
					logger.Warn("Could not perform power control action.", "action", cmdInfo.name, "error", err.Error())
				}
			}
		} else {
			callback = func(_ *paho.Publish) {
				err := bus.Call(ctx, cmdInfo.path, loginBaseInterface, cmdInfo.method, true)
				if err != nil {
					logger.Warn("Could not power off session.", "error", err.Error())
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
