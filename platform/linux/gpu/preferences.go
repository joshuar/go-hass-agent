// Copyright 2026 opticon454
// SPDX-License-Identifier: MIT

package gpu

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix = "sensors.gpu."
)

// NvidiaPrefs are the preferences for the Nvidia GPU worker.
type NvidiaPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
