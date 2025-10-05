// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix         = "sensors.network."
	loopbackDeviceName = "lo"
)

var defaultIgnoredDevices = []string{"lo", "veth", "podman", "docker", "vnet"}

// Preferences represents common preferences across all net workers. All workers support being disabled and setting a
// list of devices to filter.
type Preferences struct {
	workers.CommonWorkerPrefs

	IgnoredDevices []string `toml:"ignored_devices"`
}
