// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ui

import (
	"context"
	_ "embed"

	"github.com/joshuar/go-hass-agent/internal/tracker"
)

//go:generate moq -out mock_Agent_test.go . Agent
type Agent interface {
	IsHeadless() bool
	AppID() string
	Stop()
	GetConfig(string, any) error
	SetConfig(string, any) error
}

//go:generate moq -out mock_SensorTracker_test.go . SensorTracker
type SensorTracker interface {
	SensorList() []string
	Get(string) (tracker.Sensor, error)
}

// AgentUI are the methods required for the agent to display its windows, tray
// and notifications
//
//go:generate moq -out mockAgentUI_test.go . AgentUI
type AgentUI interface {
	DisplayNotification(string, string)
	DisplayTrayIcon(Agent, SensorTracker)
	DisplayRegistrationWindow(context.Context, Agent, chan struct{})
	Run()
}

//go:embed assets/issueURL.txt
var IssueURL string

//go:embed assets/featureRequestURL.txt
var FeatureRequestURL string

//go:embed assets/logo-pretty.png
var hassIcon []byte

type TrayIcon struct{}

func (i *TrayIcon) Name() string {
	return "TrayIcon"
}

func (i *TrayIcon) Content() []byte {
	return hassIcon
}
