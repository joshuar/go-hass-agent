// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package power

import "github.com/joshuar/go-hass-agent/internal/components/preferences"

const (
	sensorsPrefPrefix  = preferences.SensorsPrefPrefix + "power" + preferences.PathDelim
	controlsPrefPrefix = preferences.ControlsPrefPrefix + "power" + preferences.PathDelim
)
