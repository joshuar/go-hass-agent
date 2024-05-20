// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"os"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"

	"github.com/rs/zerolog/log"
)

func NewScreenLockControl(ctx context.Context) *mqtthass.ButtonEntity {
	dbusScreensaverDest, dbusScreensaverPath, dbusScreensaverMsg := getDesktopEnvScreensaverConfig()
	dbusScreensaverLockMethod := dbusScreensaverDest + ".Lock"
	device := linux.MQTTDevice()

	return mqtthass.AsButton(
		mqtthass.NewEntity(preferences.AppName, "Lock Screensaver", device.Name+"_lock_screensaver").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon("mdi:eye-lock").
			WithCommandCallback(func(_ *paho.Publish) {
				if dbusScreensaverPath == "" {
					log.Warn().Msg("Could not determine screensaver method.")
				}
				var err error
				if dbusScreensaverMsg != nil {
					err = dbusx.Call(ctx, dbusx.SessionBus, dbusScreensaverPath, dbusScreensaverDest, dbusScreensaverLockMethod, dbusScreensaverMsg)
				} else {
					err = dbusx.Call(ctx, dbusx.SessionBus, dbusScreensaverPath, dbusScreensaverDest, dbusScreensaverLockMethod)
				}
				if err != nil {
					log.Warn().Err(err).Msg("Could not lock screensaver.")
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
