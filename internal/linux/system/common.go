// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

const (
	preferencesID = "system_sensors"
)

type WorkerPrefs struct {
	HWMonUpdateInterval string `toml:"hwmon_sensor_update_interval" comment:"Time between updates of hwmon sensors (default 1m)."`
	DisableHWMon        bool   `toml:"disable_hwmon" comment:"Set to true to disable hwmon sensors."`
}
