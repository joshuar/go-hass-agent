// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	prefPrefix = preferences.SensorsPrefPrefix + "desktop" + preferences.PathDelim
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
}
