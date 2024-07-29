// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate moq -out interfaces_mocks_test.go . Registry SensorTracker
package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// UI are the methods required for the agent to display its windows, tray
// and notifications.
type UI interface {
	DisplayNotification(n ui.Notification)
	DisplayTrayIcon(ctx context.Context, agent ui.Agent, trk ui.SensorTracker)
	DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Preferences, doneCh chan struct{})
	Run(ctx context.Context, doneCh chan struct{})
}

type Registry interface {
	SetDisabled(id string, state bool) error
	SetRegistered(id string, state bool) error
	IsDisabled(id string) bool
	IsRegistered(id string) bool
}
