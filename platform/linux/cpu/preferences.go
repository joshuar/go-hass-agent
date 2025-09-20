// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix = "sensors.cpu."
)

// FreqWorkerPrefs are the preferences for the CPU frequency worker.
type FreqWorkerPrefs struct {
	UpdateInterval string `toml:"update_interval"`
	workers.CommonWorkerPrefs
}

// UsagePrefs are the preferences for the CPU usage worker.
type UsagePrefs struct {
	UpdateInterval string `toml:"update_interval"`
	workers.CommonWorkerPrefs
}
