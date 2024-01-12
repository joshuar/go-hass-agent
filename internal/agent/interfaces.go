// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/agent/ui"
)

//go:generate moq -out mockDevice.go . Device
type Device interface {
	DeviceName() string
	DeviceID() string
	Setup(ctx context.Context) context.Context
}

// UI are the methods required for the agent to display its windows, tray
// and notifications
//
//go:generate moq -out mockAgentUI_test.go . AgentUI
type UI interface {
	DisplayNotification(title, message string)
	DisplayTrayIcon(agent ui.Agent, cfg config.Config, trk ui.SensorTracker)
	DisplayRegistrationWindow(ctx context.Context, server, token *string, doneCh chan struct{})
	Run()
}
