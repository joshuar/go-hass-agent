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
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"
)

type appSensor struct {
	appList map[string]dbus.Variant
	linuxSensor
}

func newAppSensor(t sensorType, v map[string]dbus.Variant) *appSensor {
	s := &appSensor{}
	s.sensorType = t
	switch s.sensorType {
	case appRunning:
		s.icon = "mdi:apps"
		s.units = "apps"
		s.stateClass = sensor.StateMeasurement
		s.appList = v
		var count int
		for _, raw := range v {
			if appState, ok := raw.Value().(uint32); ok {
				if appState > 0 {
					count++
				}
			}
		}
		s.value = count
	case appActive:
		s.icon = "mdi:application"
		for rawName, rawValue := range v {
			if appState, ok := rawValue.Value().(uint32); ok {
				if appState == 2 {
					s.value = rawName
				}
			}
		}
	}
	return s
}

type appState struct {
	appChan    chan string
	countCh    chan int
	currentApp string
	appCount   int
}

func (s *appSensor) Attributes() interface{} {
	switch s.sensorType {
	case appActive:
		if app, ok := s.State().(string); ok {
			return newActiveAppDetails(app)
		} else {
			return nil
		}
	case appRunning:
		return newRunningAppsDetails(s.appList)
	}
	return nil
}

type activeAppDetails struct {
	// Cmd        string `json:"Command Line"`
	// Started    string `json:"Started"`
	// Count      int    `json:"Process Count"`
	DataSource string `json:"Data Source"`
}

func newActiveAppDetails(app string) *activeAppDetails {
	// TODO: profile and improve code to avoid memory leak
	// var appProcesses []*process.Process
	// var cmd string
	// var createTime int64
	// appProcesses = findProcesses(getProcessBasename(app))
	// if len(appProcesses) > 0 {
	// 	cmd, _ = appProcesses[0].Cmdline()
	// 	createTime, _ = appProcesses[0].CreateTime()
	// }
	return &activeAppDetails{
		// Cmd:        cmd,
		// Started:    time.UnixMilli(createTime).Format(time.RFC3339),
		// Count:      len(appProcesses),
		DataSource: srcDbus,
	}
}

type runningAppsDetails struct {
	DataSource  string   `json:"Data Source"`
	RunningApps []string `json:"Running Apps"`
}

func newRunningAppsDetails(apps map[string]dbus.Variant) *runningAppsDetails {
	details := new(runningAppsDetails)
	for appName, state := range apps {
		if variantToValue[uint32](state) > 0 {
			details.RunningApps = append(details.RunningApps, appName)
		}
	}
	details.DataSource = srcDbus
	return details
}

func AppUpdater(ctx context.Context, tracker device.SensorTracker) {
	portalDest := findPortal()
	if portalDest == "" {
		log.Warn().
			Msg("Unsupported or unknown portal. App sensors will not run.")
		return
	}

	appStateTracker := &appState{
		appChan: make(chan string),
		countCh: make(chan int),
	}
	go func() {
		for {
			select {
			case app := <-appStateTracker.appChan:
				appStateTracker.currentApp = app
			case count := <-appStateTracker.countCh:
				appStateTracker.appCount = count
			case <-ctx.Done():
				return
			}
		}
	}()

	err := NewBusRequest(ctx, SessionBus).
		Path(appStateDBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(appStateDBusPath),
			dbus.WithMatchInterface(appStateDBusInterface),
		}).
		Event(appStateDBusEvent).
		Handler(func(_ *dbus.Signal) {
			if activeAppList := NewBusRequest(ctx, SessionBus).
				Path(appStateDBusPath).
				Destination(portalDest).
				GetData(appStateDBusMethod).AsVariantMap(); activeAppList != nil {
				newAppCount := newAppSensor(appRunning, activeAppList)
				newApp := newAppSensor(appActive, activeAppList)
				if count, ok := newAppCount.State().(int); ok {
					if count != appStateTracker.appCount {
						appStateTracker.countCh <- count
						if err := tracker.UpdateSensors(ctx, newAppCount); err != nil {
							log.Error().Err(err).Msg("Could not update active app count.")
						}
					}
				}
				if app, ok := newApp.State().(string); ok {
					if app != appStateTracker.currentApp {
						appStateTracker.appChan <- app
						if err := tracker.UpdateSensors(ctx, newApp); err != nil {
							log.Error().Err(err).Msg("Could not update active app.")
						}
					}
				}
			} else {
				log.Debug().Caller().
					Msg("No active apps found.")
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create active app DBus watch.")
	}
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
