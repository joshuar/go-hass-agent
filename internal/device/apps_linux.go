package device

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=AppSensorType -output appSensor_types_linux.go
const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"

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

// appInfo implements the runningApps and activeApps interfaces.

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

// appState implements hass.SensorUpdate

func (a *appState) Device() string {
	return ""
}

func (a *appState) Name() string {
	return a.stateType.String()
}

func (a *appState) Icon() string {
	return "mdi:application"
}

func (a *appState) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (a *appState) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (a *appState) StateClass() hass.SensorStateClass {
	return 0
}

func (a *appState) State() interface{} {
	return a.value
}

func (a *appState) Units() string {
	return ""
}

func (a *appState) Category() string {
	return ""
}

func (a *appState) Attributes() interface{} {
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

func AppUpdater(ctx context.Context, update chan interface{}, done chan struct{}) {
	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor app state.")
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
			path: appStateDBusPath,
			intr: appStateDBusInterface,
		},
		event: appStateDBusEvent,
		eventHandler: func(s *dbus.Signal) {
			activeAppList, err := deviceAPI.GetDBusData(sessionBus,
				portalDest,
				appStateDBusPath,
				appStateDBusMethod)
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
	log.Debug().Caller().
		Msg("Adding a DBus watch for app change events.")
	deviceAPI.WatchEvents <- appChangeSignal

	<-done
	log.Debug().Caller().
		Msg("Stopping Linux app updater.")
}
