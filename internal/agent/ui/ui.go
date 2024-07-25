// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:blank-imports
package ui

import (
	_ "embed"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

//go:generate moq -out mock_Agent_test.go . Agent
type Agent interface {
	Stop()
}

//go:generate moq -out mock_SensorTracker_test.go . SensorTracker
type SensorTracker interface {
	SensorList() []string
	Get(key string) (sensor.Details, error)
}

type Notification interface {
	GetTitle() string
	GetMessage() string
}

type MQTTPreferences struct {
	Server   string
	User     string
	Password string
	Enabled  bool
}
