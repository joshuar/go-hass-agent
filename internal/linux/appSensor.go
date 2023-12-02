// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
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
		if _, ok := s.State().(string); ok {
			return newActiveAppDetails()
		} else {
			return nil
		}
	case appRunning:
		return newRunningAppsDetails(s.appList)
	}
	return nil
}

type activeAppDetails struct {
	DataSource string `json:"Data Source"`
}

func newActiveAppDetails() *activeAppDetails {
	return &activeAppDetails{
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
		if dbushelpers.VariantToValue[uint32](state) > 0 {
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

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(appStateDBusPath),
			dbus.WithMatchInterface(appStateDBusInterface),
			dbus.WithMatchMember("RunningApplicationsChanged"),
		}).
		Handler(func(_ *dbus.Signal) {
			if activeAppList := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
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
