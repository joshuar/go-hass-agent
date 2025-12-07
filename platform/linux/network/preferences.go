// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix = "sensors.network."
)

var defaultIgnoredDevices = []string{"lo", "veth", "podman", "docker", "vnet"}

// CommonPreferences represents common preferences across all net workers. All workers support being disabled and setting a
// list of devices to filter.
type CommonPreferences struct {
	workers.CommonWorkerPrefs

	IgnoredDevices []string `toml:"ignored_devices"`
}
