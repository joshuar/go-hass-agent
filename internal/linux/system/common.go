// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

const (
	preferencesID = "system_sensors"
)

type WorkerPrefs struct {
	DisableHWMon bool `toml:"disable_hwmon" comment:"Set to true to disable hwmon sensors."`
}
