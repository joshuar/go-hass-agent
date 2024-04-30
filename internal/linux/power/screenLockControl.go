// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/godbus/dbus/v5"
	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"

	"github.com/rs/zerolog/log"
)

func NewScreenLockControl(ctx context.Context) *mqtthass.EntityConfig {
	dbusScreensaverDest, dbusScreensaverPath, dbusScreensaverMsg := getDesktopEnvScreensaverConfig()
	dbusScreensaverLockMethod := dbusScreensaverDest + ".Lock"

	return linux.NewButton("lock_screensaver").
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			if dbusScreensaverPath == "" {
				log.Warn().Msg("Could not determine screensaver method.")
			}
			var err error
			if dbusScreensaverMsg != nil {
				err = sessionDBusCall(ctx, dbus.ObjectPath(dbusScreensaverPath), dbusScreensaverDest, dbusScreensaverLockMethod, dbusScreensaverMsg)
			} else {
				err = sessionDBusCall(ctx, dbus.ObjectPath(dbusScreensaverPath), dbusScreensaverDest, dbusScreensaverLockMethod)
			}
			if err != nil {
				log.Warn().Err(err).Msg("Could not lock screensaver.")
			}
		})
}

func getDesktopEnvScreensaverConfig() (dest, path string, msg *string) {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	switch {
	case strings.Contains(desktop, "KDE"):
		return "org.freedesktop.ScreenSaver", "/ScreenSaver", nil
	case strings.Contains(desktop, "GNOME"):
		return "org.gnome.ScreenSaver", "/org/gnome/ScreenSaver", nil
	case strings.Contains(desktop, "Cinnamon"):
		msg := ""
		return "org.cinnamon.ScreenSaver", "/org/cinnamon/ScreenSaver", &msg
	case strings.Contains(desktop, "XFCE"):
		msg := ""
		return "org.xfce.ScreenSaver", "/", &msg
	default:
		return "", "", nil
	}
}

func sessionDBusCall(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
	return dbusx.NewBusRequest(ctx, dbusx.SessionBus).
		Path(path).
		Destination(dest).
		Call(method, args...)
}
