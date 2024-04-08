// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	mqtthass "github.com/joshuar/go-hass-anything/v6/pkg/hass"

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

	dbusEmptyScreensaverMessage = ""
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

func newMQTTObject(ctx context.Context) *mqttObj {
	appName := "go_hass_agent"

	baseEntity := func(entityID string) *mqtthass.EntityConfig {
		return mqtthass.NewEntityByID(entityID, appName, "homeassistant").
			AsButton().
			WithDefaultOriginInfo().
			WithDeviceInfo(mqttDevice())
	}

	systemDbusCall := func(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
		return dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(path).
			Destination(dest).
			Call(method, args...)
	}

	sessionDbusCall := func(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
		return dbusx.NewBusRequest(ctx, dbusx.SessionBus).
			Path(path).
			Destination(dest).
			Call(method, args...)
	}

	dbusScreensaverDest, dbusScreensaverPath, dbusScreensaverMsg := GetDesktopEnvScreensaverConfig()
	dbusScreensaverLockMethod := dbusScreensaverDest + ".Lock"

	sessionPath := dbusx.GetSessionPath(ctx)
	entities := make(map[string]*mqtthass.EntityConfig)
	entities["lock_screensaver"] = baseEntity("lock_screensaver").
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			if dbusScreensaverPath == "" {
				log.Warn().Msg("Could not determine screensaver method.")
			}
			var err error
			if dbusScreensaverMsg != nil {
				err = sessionDbusCall(ctx, dbus.ObjectPath(dbusScreensaverPath), dbusScreensaverDest, dbusScreensaverLockMethod, dbusScreensaverMsg)
			} else {
				err = sessionDbusCall(ctx, dbus.ObjectPath(dbusScreensaverPath), dbusScreensaverDest, dbusScreensaverLockMethod)
			}
			if err != nil {
				log.Warn().Err(err).Msg("Could not lock screensaver.")
			}
		})
	for k, v := range commands {
		var callback func(MQTT.Client, MQTT.Message)
		if v.path == "" {
			callback = func(_ MQTT.Client, _ MQTT.Message) {
				err := systemDbusCall(ctx, sessionPath, dbusSessionDest, v.method)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not %s session.", v.name)
				}
			}
		} else {
			callback = func(_ MQTT.Client, _ MQTT.Message) {
				err := systemDbusCall(ctx, v.path, dbusSessionDest, v.method, true)
				if err != nil {
					log.Warn().Err(err).Msg("Could not power off session.")
				}
			}
		}
		entities[k] = baseEntity(k).
			WithIcon(v.icon).
			WithCommandCallback(callback)
	}
	return &mqttObj{
		entities: entities,
	}
}

func mqttDevice() *mqtthass.Device {
	dev := linux.NewDevice(preferences.AppName, preferences.AppVersion)
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
}

func GetDesktopEnvScreensaverConfig() (dest, path string, msg *string) {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	switch {
	case strings.Contains(desktop, "KDE"):
		return "org.freedesktop.ScreenSaver", "/ScreenSaver", nil
	case strings.Contains(desktop, "GNOME"):
		return "org.gnome.ScreenSaver", "/org/gnome/ScreenSaver", nil
	case strings.Contains(desktop, "Cinnamon"):
		msg := dbusEmptyScreensaverMessage
		return "org.cinnamon.ScreenSaver", "/org/cinnamon/ScreenSaver", &msg
	case strings.Contains(desktop, "XFCE"):
		msg := dbusEmptyScreensaverMessage
		return "org.xfce.ScreenSaver", "/", &msg
	default:
		return "", "", nil
	}
}
