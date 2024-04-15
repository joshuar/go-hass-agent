// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"context"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"

	"github.com/rs/zerolog/log"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"
)

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	portalDest := linux.FindPortal()
	if portalDest == "" {
		log.Warn().
			Msg("Unable to monitor for active applications. No app tracking available.")
		close(sensorCh)
		return sensorCh
	}
	activeApp := newActiveAppSensor()
	runningApps := newRunningAppsSensor()

	appListReq := dbusx.NewBusRequest(ctx, dbusx.SessionBus).
		Path(appStateDBusPath).
		Destination(portalDest)

	err := dbusx.NewBusRequest(ctx, dbusx.SessionBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(appStateDBusPath),
			dbus.WithMatchInterface(appStateDBusInterface),
			dbus.WithMatchMember("RunningApplicationsChanged"),
		}).
		Handler(func(_ *dbus.Signal) {
			appList, err := dbusx.GetData[map[string]dbus.Variant](appListReq, appStateDBusMethod)
			if err != nil {
				log.Warn().Err(err).Msg("Could not retrieve app list from D-Bus.")
			}
			if appList != nil {
				activeApp.update(appList, sensorCh)
				runningApps.update(appList, sensorCh)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create active app D-Bus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped app sensor.")
	}()
	return sensorCh
}
