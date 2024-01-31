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

	mqtthass "github.com/joshuar/go-hass-anything/v3/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusSessionDest    = "org.freedesktop.login1"
	dbusSessionLockMethod     = dbusSessionDest + ".Session.Lock"
	dbusSessionUnlockMethod   = dbusSessionDest + ".Session.UnLock"
	dbusSessionRebootMethod   = dbusSessionDest + ".Manager.Reboot"
	dbusSessionPowerOffMethod = dbusSessionDest + ".Manager.PowerOff"

	dbusEmptyScreensaverMessage = ""
)

func newMQTTObject(ctx context.Context) *mqttObj {
	appName := "go_hass_agent"

	baseEntity := func(entityID string) *mqtthass.EntityConfig {
		return mqtthass.NewEntityByID(entityID, appName).
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

	sessionPath := dbusx.GetSessionPath(ctx)
	entities := make(map[string]*mqtthass.EntityConfig)
	entities["lock_screensaver"] = baseEntity("lock_screensaver").
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			dbusScreensaverDest, dbusScreensaverPath, dbusScreensaverMsg := GetDesktopEnvScreensaverConfig();
			if dbusScreensaverPath == "" {
				log.Warn().Msg("Could not determine screensaver method.")
				return
			}
			dbusScreensaverLockMethod := dbusScreensaverDest + ".Lock"
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
	entities["lock_session"] = baseEntity("lock_session").
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := systemDbusCall(ctx, sessionPath, dbusSessionDest, dbusSessionLockMethod)
			if err != nil {
				log.Warn().Err(err).Msg("Could not lock session.")
			}
		})
	entities["unlock_session"] = baseEntity("unlock_session").
		WithIcon("mdi:eye-lock-open").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := systemDbusCall(ctx, sessionPath, dbusSessionDest, dbusSessionUnlockMethod)
			if err != nil {
				log.Warn().Err(err).Msg("Could not unlock session.")
			}
		})
	entities["reboot"] = baseEntity("reboot").
		WithIcon("mdi:restart").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := systemDbusCall(ctx, dbus.ObjectPath("/org/freedesktop/login1"), dbusSessionDest, dbusSessionRebootMethod, true)
			if err != nil {
				log.Warn().Err(err).Msg("Could not reboot session.")
			}
		})
	entities["poweroff"] = baseEntity("poweroff").
		WithIcon("mdi:power").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := systemDbusCall(ctx, dbus.ObjectPath("/org/freedesktop/login1"), dbusSessionDest, dbusSessionPowerOffMethod, true)
			if err != nil {
				log.Warn().Err(err).Msg("Could not power off session.")
			}
		})
	return &mqttObj{
		entities: entities,
	}
}

func mqttDevice() *mqtthass.Device {
	dev := linux.NewDevice(config.AppName, config.AppVersion)
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          config.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
}

func GetDesktopEnvScreensaverConfig() (string, string, *string) {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	switch {
	case strings.Contains(desktop, "KDE"):
		return "org.freedesktop.ScreenSaver", "/ScreenSaver", nil
	case strings.Contains(desktop, "GNOME"):
		return "org.gnome.ScreenSaver", "/org/gnome/ScreenSaver", nil
	case strings.Contains(desktop, "Cinnamon"):
		msg := dbusEmptyScreensaverMessage
		return "org.cinnamon.ScreenSaver", "/org/cinnamon/ScreenSaver", &msg
	default:
		return "", "", nil
	}
}
