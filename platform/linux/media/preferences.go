// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import "github.com/joshuar/go-hass-agent/agent/workers"

const (
	prefPrefix  = "sensors.media."
	mprisPrefID = prefPrefix + "mpris"
)

type WorkerPrefs struct {
	*workers.CommonWorkerPrefs

	UpdateInterval string `toml:"update_interval"`
}
