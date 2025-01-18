// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

import (
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

const (
	preferencesID      = "network_sensors"
	loopbackDeviceName = "lo"
)

var defaultIgnoredDevices = []string{}

//nolint:lll
type WorkerPrefs struct {
	IgnoredDevices []string `toml:"ignored_devices" comment:"List of prefixes to match for devices to ignore, for e.g., 'eth' will ignore all devices starting with eth."`
	UpdateInterval string   `toml:"network_stats_update_interval" comment:"Time between updates of network stats sensors."`
	preferences.CommonWorkerPrefs
}
