package device

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=AppSensorType -output appSensor_types_linux.go
const (
	appStateDBusMethod = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath   = "/org/freedesktop/portal/desktop"

	ActiveApp AppSensorType = iota
	RunningApps
)

type AppSensorType int

type activeApp interface {
	Name() string
	Attributes() interface{}
}

type runningApps interface {
	Count() int
	Attributes() interface{}
}

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
		RunningApps []string `json:"Running Apps"`
	}{
		RunningApps: runningApps,
	}
}

type appState struct {
	value      interface{}
	stateType  AppSensorType
	attributes interface{}
}

func (a *appState) ID() string {
	return ""
}

func (a *appState) Type() string {
	return a.stateType.String()
}

func (a *appState) Value() interface{} {
	return a.value
}

func (a *appState) ExtraValues() interface{} {
	return a.attributes
}

func (a *appInfo) marshallStateUpdate(t AppSensorType) *appState {
	switch t {
	case RunningApps:
		return &appState{
			value:      a.Count(),
			stateType:  t,
			attributes: a.Attributes(),
		}
	case ActiveApp:
		return &appState{
			value:      a.Name(),
			stateType:  t,
			attributes: a.Attributes(),
		}
	default:
		return nil
	}
}

func AppUpdater(ctx context.Context, update chan interface{}) {

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
				update <- a.marshallStateUpdate(RunningApps)
				update <- a.marshallStateUpdate(ActiveApp)
			}
		},
	}
	log.Debug().Caller().Msg("Adding a DBus watch for app change events.")
	deviceAPI.WatchEvents <- appChangeSignal

	<-update
	// for {
	// 	select {
	// 	case <-ctx.Done():
	log.Debug().Caller().
		Msg("Stopping Linux app updater.")
	// return
	// 	}
	// }
}
