// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	basePreferencesID        = "disk_sensors"
	ioWorkerPreferencesID    = "io_sensors"
	usageWorkerPreferencesID = "usage_sensors"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of sensors."`
}
