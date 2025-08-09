// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	preferencesID = preferences.SensorsPrefPrefix + "batteries"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
}
