// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	ioWorkerPreferencesID    = "disk_io_sensors"
	usageWorkerPreferencesID = "disk_usage_sensors"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of sensors."`
}
