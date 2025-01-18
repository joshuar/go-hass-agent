// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	preferencesID = "battery_sensors"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
}
