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

var defaultIgnoredDevices = []string{"veth", "podman", "docker", "vnet"}

//nolint:lll
type WorkerPrefs struct {
	workers.CommonWorkerPrefs

	IgnoredDevices []string `toml:"ignored_devices"`
}
