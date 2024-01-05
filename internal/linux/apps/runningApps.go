// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
)

type runningAppsSensor struct {
	appList map[string]dbus.Variant
	mu      sync.Mutex
	linux.Sensor
}

type runningAppsSensorAttributes struct {
	DataSource  string   `json:"Data Source"`
	RunningApps []string `json:"Running Apps"`
}

func (r *runningAppsSensor) Attributes() any {
	attrs := &runningAppsSensorAttributes{}
	r.mu.Lock()
	for appName, state := range r.appList {
		if dbushelpers.VariantToValue[uint32](state) > 0 {
			attrs.RunningApps = append(attrs.RunningApps, appName)
		}
	}
	r.mu.Unlock()
	attrs.DataSource = linux.DataSrcDbus
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
	r.mu.Lock()
	r.appList = l
	for _, raw := range l {
		if appState, ok := raw.Value().(uint32); ok {
			if appState > 0 {
				count++
			}
		}
	}
	r.mu.Unlock()
	if r.count() != count {
		r.Value = count
		s <- r
	}
}

func newRunningAppsSensor() *runningAppsSensor {
	s := &runningAppsSensor{}
	s.SensorTypeValue = linux.SensorAppRunning
	s.IconString = "mdi:apps"
	s.UnitsString = "apps"
	s.StateClassValue = sensor.StateMeasurement
	return s
}
