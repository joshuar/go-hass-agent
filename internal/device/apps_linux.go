package device

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	appStateDBusMethod = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath   = "/org/freedesktop/portal/desktop"
)

type appInfo struct {
	activeApps map[string]interface{}
}

func (a *appInfo) Name() string {
	for key, value := range a.activeApps {
		if value.(uint32) == 2 {
			return key
		}
	}
	return "Unknown"
}

func (a *appInfo) Count() int {
	var count int
	for _, value := range a.activeApps {
		if value.(uint32) > 0 {
			count++
		}
	}
	return count
}

func (a *appInfo) Attributes() interface{} {
	var runningApps []string
	for key, value := range a.activeApps {
		if value.(uint32) > 0 {
			runningApps = append(runningApps, key)
		}
	}
	return struct {
		RunningApps []string `json:"running_apps"`
	}{
		RunningApps: runningApps,
	}
}

func AppUpdater(ctx context.Context, app chan interface{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	portalDest := FindPortal()
	if portalDest == "" {
		log.Debug().Caller().
			Msgf("Unsupported or unknown portal")
		return
	}

	a := &appInfo{}
	appChangeSignal := &DBusWatchRequest{
		bus: sessionBus,
		match: DBusSignalMatch{
			path: "/org/freedesktop/portal/desktop",
			intr: "org.freedesktop.impl.portal.Background",
		},
		event: "org.freedesktop.impl.portal.Background.RunningApplicationsChanged",
		eventHandler: func(s *dbus.Signal) {
			activeAppList, err := deviceAPI.GetDBusData(sessionBus, portalDest, appStateDBusPath, appStateDBusMethod)
			if err != nil {
				log.Debug().Caller().Msgf(err.Error())
			} else {
				a.activeApps = nil
				a.activeApps = activeAppList.(map[string]interface{})
				app <- a
			}
		},
	}
	log.Debug().Caller().Msg("Adding a DBus watch for app change events.")
	deviceAPI.WatchEvents <- appChangeSignal

	for {
		select {
		// case <-c:
		// 	obj := deviceAPI.dBusSession.Object(portalDest, appStateDBusPath)
		// 	var activeAppList map[string]interface{}
		// 	err = obj.Call(appStateDBusMethod, 0).Store(&activeAppList)
		// 	if err != nil {
		// 		log.Debug().Caller().Msgf(err.Error())
		// 	} else {
		// 		a.activeApps = nil
		// 		a.activeApps = activeAppList
		// 		app <- a
		// 	}
		case <-ctx.Done():
		case <-app:
			log.Debug().Caller().
				Msg("Stopping Linux app updater.")
			return
		}
	}
}
