// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix = "sensors.cpu."
)

// FreqPrefs are the preferences for the CPU frequency worker.
type FreqPrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}

// UsagePrefs are the preferences for the CPU usage worker.
type UsagePrefs struct {
	workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}
