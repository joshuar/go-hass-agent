// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	prefPrefix        = preferences.SensorsPrefPrefix + "media" + preferences.PathDelim
	mprisPrefID       = prefPrefix + "mpris"
	webcamUsagePrefID = prefPrefix + "webcam_in_use"
	micUsagePrefID    = prefPrefix + "microphone_in_use"
)

type WorkerPrefs struct {
	preferences.CommonWorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of sensors."`
}
