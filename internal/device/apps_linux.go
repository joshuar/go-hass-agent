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
	monitorConn, err := DBusConnectSession(ctx)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to DBus to monitor app state: %v", err)
		return
	}
	defer monitorConn.Close()

	rules := []string{
		"type='signal',member='RunningApplicationsChanged',path='/org/freedesktop/portal/desktop',interface='org.freedesktop.impl.portal.Background'",
	}
	err = DBusBecomeMonitor(monitorConn, rules, 0)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could become monitor: %v", err)
		return
	}

	c := make(chan *dbus.Message, 10)
	defer close(c)
	monitorConn.Eavesdrop(c)
	log.Debug().Caller().Msg("Monitoring DBus for app changes.")

	appChkConn, err := DBusConnectSession(ctx)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to DBus to monitor app state: %v", err)
		return
	}

	portalDest := FindPortal()
	if portalDest == "" {
		log.Debug().Caller().
			Msgf("Unsupported or unknown portal")
		return
	}

	a := &appInfo{}

	for {
		select {
		case <-c:
			obj := appChkConn.Object(portalDest, appStateDBusPath)
			var activeAppList map[string]interface{}
			err = obj.Call(appStateDBusMethod, 0).Store(&activeAppList)
			if err != nil {
				log.Debug().Caller().Msgf(err.Error())
			} else {
				a.activeApps = nil
				a.activeApps = activeAppList
				app <- a
			}
		case <-ctx.Done():
		case <-app:
			log.Debug().Caller().
				Msg("Stopping Linux app updater.")
			return
		}
	}
}
