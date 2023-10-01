// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
)

// AgentConfig represents the methods that the agent uses to interact with
// its config. It is effectively a CRUD interface to wherever the configuration
// is stored.
//
//go:generate moq -out mockAgentConfig_test.go . AgentConfig
type AgentConfig interface {
	Get(string, interface{}) error
	Set(string, interface{}) error
	Delete(string) error
	StoragePath(string) (string, error)
}

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
