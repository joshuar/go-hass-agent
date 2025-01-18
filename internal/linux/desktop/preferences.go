// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	preferencesID = "desktop_sensors"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
}
