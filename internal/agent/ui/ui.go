// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ui

import (
	_ "embed"

	"github.com/joshuar/go-hass-agent/internal/tracker"
)

//go:generate moq -out mock_Agent_test.go . Agent
type Agent interface {
	AppID() string
	Stop()
}

type AgentOptions interface {
	GetServer() string
	GetToken() string
	SetServer(value string)
	SetToken(value string)
}

//go:generate moq -out mock_SensorTracker_test.go . SensorTracker
type SensorTracker interface {
	SensorList() []string
	Get(key string) (tracker.Sensor, error)
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
