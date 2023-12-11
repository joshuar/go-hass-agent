// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"
)

type runningAppsSensor struct {
	appList map[string]dbus.Variant
	linuxSensor
}

type runningAppsSensorAttributes struct {
	DataSource  string   `json:"Data Source"`
	RunningApps []string `json:"Running Apps"`
}

func (r *runningAppsSensor) Attributes() any {
	attrs := &runningAppsSensorAttributes{}
	for appName, state := range r.appList {
		if dbushelpers.VariantToValue[uint32](state) > 0 {
			attrs.RunningApps = append(attrs.RunningApps, appName)
		}
	}
	attrs.DataSource = srcDbus
	return attrs
}

func (r *runningAppsSensor) count() int {
	if count, ok := r.State().(int); ok {
		return count
	}
	return -1
}

func (r *runningAppsSensor) update(l map[string]dbus.Variant, s chan tracker.Sensor) {
	var count int
	r.appList = l
	for _, raw := range l {
		if appState, ok := raw.Value().(uint32); ok {
			if appState > 0 {
				count++
			}
		}
	}
	if r.count() != count {
		r.value = count
		s <- r
	}
}

func newRunningAppsSensor() *runningAppsSensor {
	s := &runningAppsSensor{}
	s.sensorType = appRunning
	s.icon = "mdi:apps"
	s.units = "apps"
	s.stateClass = sensor.StateMeasurement
	return s
}

type activeAppSensor struct {
	linuxSensor
}

type activeAppSensorAttributes struct {
	DataSource string `json:"Data Source"`
}

func (a *activeAppSensor) Attributes() any {
	if _, ok := a.value.(string); ok {
		return &activeAppSensorAttributes{
			DataSource: srcDbus,
		}
	}
	return nil
}

func (a *activeAppSensor) app() string {
	if app, ok := a.State().(string); ok {
		return app
	}
	return ""
}

func (a *activeAppSensor) update(l map[string]dbus.Variant, s chan tracker.Sensor) {
	for app, v := range l {
		if appState, ok := v.Value().(uint32); ok {
			if appState == 2 && a.app() != app {
				a.value = app
				s <- a
			}
		}
	}
}

func newActiveAppSensor() *activeAppSensor {
	s := &activeAppSensor{}
	s.sensorType = appActive
	s.icon = "mdi:application"
	return s
}

func AppUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	portalDest := findPortal()
	if portalDest == "" {
		log.Warn().
			Msg("Unsupported or unknown portal. App sensors will not run.")
		close(sensorCh)
		return sensorCh
	}
	activeApp := newActiveAppSensor()
	runningApps := newRunningAppsSensor()

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(appStateDBusPath),
			dbus.WithMatchInterface(appStateDBusInterface),
			dbus.WithMatchMember("RunningApplicationsChanged"),
		}).
		Handler(func(_ *dbus.Signal) {
			appList := dbushelpers.NewBusRequest(ctx, dbushelpers.SessionBus).
				Path(appStateDBusPath).
				Destination(portalDest).
				GetData(appStateDBusMethod).AsVariantMap()
			if appList != nil {
				activeApp.update(appList, sensorCh)
				runningApps.update(appList, sensorCh)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create active app DBus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped app sensor.")
	}()
	return sensorCh
}
