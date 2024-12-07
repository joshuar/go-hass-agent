// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	preferencesID = "cpu_sensors"
)

type WorkerPrefs struct {
	UpdateInterval string `toml:"sensor_update_interval" comment:"Time between updates of sensors (default 10s)."`
	DisableCPUFreq bool   `toml:"disable_cpufreq" comment:"Set to true to disable CPU frequency sensors."`
	preferences.CommonWorkerPrefs
}
