// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import "github.com/joshuar/go-hass-agent/agent/workers"

const (
	sensorsPrefPrefix  = "sensors.system."
	controlsPrefPrefix = "controls.system."
)

// HWMonPrefs are the preferences for the hwmon sensor worker.
type HWMonPrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}

// ProblemsPrefs are the preferences for the abrt problems sensor worker.
type ProblemsPrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}

// ChronyPrefs are the preferences for the chrony sensor worker.
type ChronyPrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}

// UptimePrefs are the preferences for the system uptime sensor.
type UptimePrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}

// UserSessionsPrefs are the preferences for the user sessions worker.
type UserSessionsPrefs struct {
	workers.CommonWorkerPrefs
}
