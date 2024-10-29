// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

const (
	preferencesID      = "network_sensors"
	loopbackDeviceName = "lo"
)

var defaultIgnoredDevices = []string{}

//nolint:lll
type WorkerPrefs struct {
	IgnoredDevices []string `toml:"ignored_devices" comment:"list of prefixes to match for devices to ignore, for e.g., 'eth' will ignore all devices starting with eth"`
}
