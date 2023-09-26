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
	IsHeadless() bool
	AppVersion() string
	AppName() string
	AppID() string
	Stop()
	GetConfig(string, interface{}) error
	SetConfig(string, interface{}) error
	SensorList() []string
	SensorValue(string) (tracker.Sensor, error)
}

//go:embed assets/issueURL.txt
var issueURL string

//go:embed assets/featureRequestURL.txt
var featureRequestURL string

//go:embed assets/logo-pretty.png
var hassIcon []byte

type TrayIcon struct{}

func (icon *TrayIcon) Name() string {
	return "TrayIcon"
}

func (icon *TrayIcon) Content() []byte {
	return hassIcon
}
