// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

//nolint:lll
func NewScreenLockControl(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger, device *mqtthass.Device) *mqtthass.ButtonEntity {
	logger := parentLogger.With(slog.String("controller", "screen_lock"))

	bus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		logger.Warn("Cannot create screen lock control.", "error", err.Error())

		return nil
	}

	dbusScreensaverDest, dbusScreensaverPath, dbusScreensaverMsg := getDesktopEnvScreensaverConfig()
	dbusScreensaverLockMethod := dbusScreensaverDest + ".Lock"

	return mqtthass.AsButton(
		mqtthass.NewEntity(preferences.AppName, "Lock Screensaver", device.Name+"_lock_screensaver").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon("mdi:eye-lock").
			WithCommandCallback(func(_ *paho.Publish) {
				if dbusScreensaverPath == "" {
					logger.Warn("Could not determine D-Bus method to control screensaver.")
				}

				var err error

				if dbusScreensaverMsg != nil {
					err = bus.Call(ctx, dbusScreensaverPath, dbusScreensaverDest, dbusScreensaverLockMethod, dbusScreensaverMsg)
				} else {
					err = bus.Call(ctx, dbusScreensaverPath, dbusScreensaverDest, dbusScreensaverLockMethod)
				}

				if err != nil {
					logger.Warn("Could not toggle screensaver.", "error", err.Error())
				}
			}))
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
