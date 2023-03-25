package device

import (
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

var ignoreApps = []string{"org.kde.plasmashell"}

type appInfo struct {
	activeApps map[string]interface{}
	// sensors    map[string]*hass.SensorInfo
}

func (a *appInfo) Name() string {
	for key, value := range a.activeApps {
		if value.(uint32) == 2 {
			if !slices.Contains(ignoreApps, key) {
				return key
			}
		}
	}
	return "Unknown"
}

func ActiveAppUpdater(app chan interface{}) {
	monitorConn, err := dbus.ConnectSessionBus()
	logging.CheckError(err)
	defer monitorConn.Close()

	rules := []string{
		"type='signal',member='RunningApplicationsChanged',path='/org/freedesktop/portal/desktop',interface='org.freedesktop.impl.portal.Background'",
	}
	var flag uint = 0

	call := monitorConn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, flag)
	logging.CheckError(call.Err)

	c := make(chan *dbus.Message, 10)
	monitorConn.Eavesdrop(c)
	log.Debug().Caller().Msg("Monitoring DBUS for active app changes")

	appChkConn, err := dbus.ConnectSessionBus()
	logging.CheckError(err)

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

	for range c {
		obj := appChkConn.Object(portalDest, "/org/freedesktop/portal/desktop")
		err = obj.Call("org.freedesktop.impl.portal.Background.GetAppState", 0).Store(&a.activeApps)
		logging.CheckError(err)
		log.Debug().Caller().Msgf("Current active app: %s.", a.Name())
		app <- a
	}

}
