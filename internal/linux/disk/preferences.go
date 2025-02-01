// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	prefPrefix               = preferences.SensorsPrefPrefix + "disk" + preferences.PathDelim
	ioWorkerPreferencesID    = prefPrefix + "rates"
	usageWorkerPreferencesID = prefPrefix + "usage"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of sensors."`
}
