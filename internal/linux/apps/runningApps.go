// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type runningAppsSensor struct {
	appList map[string]dbus.Variant
	linux.Sensor
	mu sync.Mutex
}

type runningAppsSensorAttributes struct {
	DataSource  string   `json:"data_source"`
	RunningApps []string `json:"running_apps"`
}

//nolint:exhaustruct
func (r *runningAppsSensor) Attributes() any {
	attrs := &runningAppsSensorAttributes{}

	r.mu.Lock()
	for appName, state := range r.appList {
		if dbusx.VariantToValue[uint32](state) > 0 {
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

func (r *runningAppsSensor) update(apps map[string]dbus.Variant) sensor.Details {
	var count int

	r.mu.Lock()
	defer r.mu.Unlock()
	r.appList = apps

	for _, appState := range apps {
		if appState, ok := appState.Value().(uint32); ok {
			if appState > 0 {
				count++
			}
		}
	}

	if r.count() != count {
		r.Value = count

		return r
	}

	return nil
}

//nolint:exhaustruct
func newRunningAppsSensor() *runningAppsSensor {
	newSensor := &runningAppsSensor{}
	newSensor.SensorTypeValue = linux.SensorAppRunning
	newSensor.IconString = "mdi:apps"
	newSensor.UnitsString = "apps"
	newSensor.StateClassValue = types.StateClassMeasurement

	return newSensor
}
