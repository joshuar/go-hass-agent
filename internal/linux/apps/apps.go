// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"context"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"

	"github.com/rs/zerolog/log"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"
)

func Updater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	portalDest := linux.FindPortal()
	if portalDest == "" {
		log.Warn().
			Msg("Unsupported or unknown portal. App sensors will not run.")
		close(sensorCh)
		return sensorCh
	}
	activeApp := newActiveAppSensor()
	runningApps := newRunningAppsSensor()

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(appStateDBusPath),
			dbus.WithMatchInterface(appStateDBusInterface),
			dbus.WithMatchMember("RunningApplicationsChanged"),
		}).
		Handler(func(_ *dbus.Signal) {
			appList := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
				Path(appStateDBusPath).
				Destination(portalDest).
				GetData(appStateDBusMethod).AsVariantMap()
			if appList != nil {
				activeApp.update(appList, sensorCh)
				runningApps.update(appList, sensorCh)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create active app DBus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped app sensor.")
	}()
	return sensorCh
}
