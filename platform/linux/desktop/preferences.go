// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import (
	"github.com/joshuar/go-hass-agent/agent/workers"
)

const (
	prefPrefix = "sensors.desktop."
)

type WorkerPrefs struct {
	*workers.CommonWorkerPrefs
}
