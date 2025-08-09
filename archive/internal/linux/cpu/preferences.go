// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	prefPrefix = preferences.SensorsPrefPrefix + "cpu" + preferences.PathDelim
)

// FreqWorkerPrefs are the preferences for the CPU frequency worker.
type FreqWorkerPrefs struct {
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of CPU frequency sensors (default 30s)."`
	preferences.CommonWorkerPrefs
}

// UsagePrefs are the preferences for the CPU usage worker.
type UsagePrefs struct {
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of CPU usage sensors (default 10s)."`
	preferences.CommonWorkerPrefs
}
