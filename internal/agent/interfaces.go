// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

//go:generate moq -out mockDevice_test.go . Device
type Device interface {
	DeviceName() string
	DeviceID() string
	Setup(ctx context.Context) context.Context
	Updates() chan sensor.Details
}

// UI are the methods required for the agent to display its windows, tray
// and notifications
//
//go:generate moq -out mockUI_test.go . UI
type UI interface {
	DisplayNotification(n ui.Notification)
	DisplayTrayIcon(ctx context.Context, agent ui.Agent, trk ui.SensorTracker)
	DisplayRegistrationWindow(ctx context.Context, input *hass.RegistrationInput, doneCh chan struct{})
	Run(ctx context.Context, doneCh chan struct{})
}

//go:generate moq -out mockSensorTracker_test.go . SensorTracker
type SensorTracker interface {
	SensorList() []string
	Process(ctx context.Context, reg sensor.Registry, sensorUpdates ...<-chan sensor.Details) error
	Get(key string) (sensor.Details, error)
	Reset()
}
