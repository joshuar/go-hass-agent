// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import "github.com/joshuar/go-hass-agent/agent/workers"

const (
	prefPrefix               = "sensors.disk."
	ioWorkerPreferencesID    = prefPrefix + "rates"
	usageWorkerPreferencesID = prefPrefix + "usage"
	smartWorkerPreferencesID = prefPrefix + "smart"
)

type WorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
