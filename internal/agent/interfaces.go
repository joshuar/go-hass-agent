// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
)


// AgentUI are the methods required for the agent to display its windows, tray
// and notifications
//
//go:generate moq -out mockAgentUI_test.go . AgentUI
type AgentUI interface {
	DisplayNotification(string, string)
	DisplayTrayIcon(context.Context, ui.Agent)
	DisplayRegistrationWindow(context.Context, ui.Agent, chan struct{})
	Run()
}

//go:generate moq -out mockDevice.go . Device
type Device interface {
	DeviceName() string
	DeviceID() string
	Setup(context.Context) context.Context
}
