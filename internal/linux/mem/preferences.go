// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mem

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	prefPrefix = preferences.SensorsPrefPrefix + "memory" + preferences.PathDelim
)

type WorkerPreferences struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of sensors."`
}
