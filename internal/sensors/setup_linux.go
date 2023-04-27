// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func SetupSensors() *device.SensorInfo {
	sensorInfo := device.NewSensorInfo()
	sensorInfo.Add("Location", linux.LocationUpdater)
	sensorInfo.Add("Battery", linux.BatteryUpdater)
	sensorInfo.Add("Apps", linux.AppUpdater)
	sensorInfo.Add("Network", linux.NetworkUpdater)
	sensorInfo.Add("Power", linux.PowerUpater)
	sensorInfo.Add("ExternalIP", device.ExternalIPUpdater)
	sensorInfo.Add("Problems", linux.ProblemsUpdater)
	sensorInfo.Add("Memory", linux.MemoryUpdater)
	sensorInfo.Add("LoadAvg", linux.LoadAvgUpdater)
	sensorInfo.Add("DiskUsage", linux.DiskUsageUpdater)
	// Add each SensorUpdater function here...
	return sensorInfo
}
