// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mem

import "github.com/joshuar/go-hass-agent/agent/workers"

const (
	prefPrefix = "sensors.memory."
)

type WorkerPreferences struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
