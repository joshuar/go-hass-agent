// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	sensorsPrefPrefix  = preferences.SensorsPrefPrefix + "system" + preferences.PathDelim
	controlsPrefPrefix = preferences.ControlsPrefPrefix + "system" + preferences.PathDelim
)

// HWMonPrefs are the preferences for the hwmon sensor worker.
type HWMonPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of hwmon sensors (default 1m)."`
}

// ProblemsPrefs are the preferences for the abrt problems sensor worker.
type ProblemsPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between checking for new problems."`
}

// ChronyPrefs are the preferences for the chrony sensor worker.
type ChronyPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between checking chrony status."`
}

// UptimePrefs are the preferences for the system uptime sensor.
type UptimePrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between sending uptime sensor."`
}

// UserSessionsPrefs are the preferences for the user sessions worker.
type UserSessionsPrefs struct {
	preferences.CommonWorkerPrefs
}
