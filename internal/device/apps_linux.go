package device

import (
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/rs/zerolog/log"
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

func AppUpdater(app chan interface{}) {
	monitorConn, err := ConnectSessionBus()
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to DBus to monitor app state: %v", err)
		return

	}
	defer monitorConn.Close()

	rules := []string{
		"type='signal',member='RunningApplicationsChanged',path='/org/freedesktop/portal/desktop',interface='org.freedesktop.impl.portal.Background'",
	}
	var flag uint = 0

	call := monitorConn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, flag)
	logging.CheckError(call.Err)

	c := make(chan *dbus.Message, 10)
	monitorConn.Eavesdrop(c)
	log.Debug().Caller().Msg("Monitoring D-Bus for app changes.")

	appChkConn, err := ConnectSessionBus()
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to DBus to monitor app state: %v", err)
		return

	}

	var portalDest string
	switch os.Getenv("XDG_CURRENT_DESKTOP") {
	case "KDE":
		portalDest = "org.freedesktop.impl.portal.desktop.kde"
	case "GNOME":
		portalDest = "org.freedesktop.impl.portal.desktop.kde"
	default:
		log.Warn().Msg("Unsupported desktop/window environment. No app logging available.")
	}

	a := &appInfo{}

	for {
		select {
		case <-c:
			// for range c {
			obj := appChkConn.Object(portalDest, "/org/freedesktop/portal/desktop")
			var activeAppList map[string]interface{}
			err = obj.Call("org.freedesktop.impl.portal.Background.GetAppState", 0).Store(&activeAppList)
			if err != nil {
				log.Warn().Msgf("Problem getting active app list: %v.", err)
			} else {
				a.activeApps = nil
				a.activeApps = activeAppList
				app <- a
			}
		case <-app:
			log.Debug().Caller().
				Msg("Stopping Linux app updater.")
			monitorConn.Close()
			appChkConn.Close()
			return
		}
	}
}
