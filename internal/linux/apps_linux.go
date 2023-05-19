// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/process"
)

//go:generate stringer -type=appSensorType -output appSensor_types_linux.go -linecomment
const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"

	activeApp   appSensorType = iota // Active App
	runningApps                      // Running Apps
)

type appSensorType int

type appSensor struct {
	sensorType  appSensorType
	sensorValue map[string]dbus.Variant
}

// appSensor implements hass.SensorUpdate

func (s *appSensor) Name() string {
	return s.sensorType.String()
}

func (s *appSensor) ID() string {
	return strings.ToLower(strcase.ToSnake(s.sensorType.String()))
}

func (s *appSensor) Icon() string {
	switch s.sensorType {
	case runningApps:
		return "mdi:apps"
	case activeApp:
		fallthrough
	default:
		return "mdi:application"
	}
}

func (s *appSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (s *appSensor) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (s *appSensor) StateClass() hass.SensorStateClass {
	switch s.sensorType {
	case runningApps:
		return hass.StateMeasurement
	default:
		return 0
	}
}

func (s *appSensor) State() interface{} {
	switch s.sensorType {
	case activeApp:
		for appName, state := range s.sensorValue {
			if state.Value().(uint32) == 2 {
				return appName
			}
		}
	case runningApps:
		var count int
		for _, state := range s.sensorValue {
			if state.Value().(uint32) > 0 {
				count++
			}
		}
		return count
	}
	return ""
}

func (s *appSensor) Units() string {
	switch s.sensorType {
	case runningApps:
		return "apps"
	}
	return ""
}

func (s *appSensor) Category() string {
	return ""
}

func (s *appSensor) Attributes() interface{} {
	switch s.sensorType {
	case activeApp:
		appProcesses := findProcesses(getProcessBasename(s.State().(string)))
		var cmd string
		var createTime int64
		if len(appProcesses) > 0 {
			cmd, _ = appProcesses[0].Cmdline()
			createTime, _ = appProcesses[0].CreateTime()
		}
		return struct {
			Cmd     string `json:"Command Line"`
			Count   int    `json:"Process Count"`
			Started string `json:"Started"`
		}{
			Cmd:     cmd,
			Count:   len(appProcesses),
			Started: time.UnixMilli(createTime).Format(time.RFC3339),
		}
	case runningApps:
		var runningApps []string
		for appName, state := range s.sensorValue {
			if variantToValue[uint32](state) > 0 {
				runningApps = append(runningApps, appName)
			}
		}
		return struct {
			RunningApps []string `json:"Running Apps"`
		}{
			RunningApps: runningApps,
		}
	}
	return nil
}

func marshalAppStateUpdate(t appSensorType, v map[string]dbus.Variant) *appSensor {
	return &appSensor{
		sensorValue: v,
		sensorType:  t,
	}
}

func AppUpdater(ctx context.Context, update chan interface{}) {
	deviceAPI, err := device.FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}
	dbusAPI := deviceAPI.EndPoint("session").(*bus)

	portalDest := findPortal()
	if portalDest == "" {
		log.Debug().Caller().
			Msgf("Unsupported or unknown portal")
		return
	}

	appStateDBusMatches := []dbus.MatchOption{
		dbus.WithMatchObjectPath(appStateDBusPath),
		dbus.WithMatchInterface(appStateDBusInterface),
	}
	appStateHandler := func(s *dbus.Signal) {
		activeAppList := NewBusRequest(dbusAPI).
			Path(appStateDBusPath).
			Destination(portalDest).
			GetData(appStateDBusMethod).AsVariantMap()
		if activeAppList == nil {
			log.Debug().Err(err).Caller().
				Msg("No active apps found.")
		} else {
			update <- marshalAppStateUpdate(runningApps, activeAppList)
			update <- marshalAppStateUpdate(activeApp, activeAppList)
		}
	}
	NewBusRequest(dbusAPI).
		Path(appStateDBusPath).
		Match(appStateDBusMatches).
		Event(appStateDBusEvent).
		Handler(appStateHandler).
		AddWatch(ctx)
}

func findProcesses(name string) []*process.Process {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	allProcesses, err := process.ProcessesWithContext(ctx)
	if err != nil {
		log.Debug().Caller().
			Msg("Unable to retrieve processes list.")
		cancel()
		return nil
	}
	var matchedProcesses []*process.Process
	for _, p := range allProcesses {
		if n, _ := p.Name(); strings.Contains(n, name) {
			matchedProcesses = append(matchedProcesses, p)
		}
	}
	return matchedProcesses
}

func getProcessBasename(name string) string {
	s := strings.Split(name, ".")
	return s[len(s)-1]
}
