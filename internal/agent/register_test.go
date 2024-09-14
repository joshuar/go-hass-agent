// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestAgent_checkRegistration(t *testing.T) {
	ctx := preferences.AppIDToContext(context.TODO(), "go-hass-agent-test")

	type args struct {
		ctx     context.Context
		agentUI ui
		prefs   agentPreferences
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "already registered",
			args: args{
				ctx: LoadCtx(ctx, SetForceRegister(false)),
				prefs: &agentPreferencesMock{
					AgentRegisteredFunc:     func() bool { return true },
					GetDeviceInfoFunc:       func() *preferences.Device { return &preferences.Device{} },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
			},
		},
		{
			name: "force registration",
			args: args{
				ctx: LoadCtx(ctx,
					SetHeadless(false),
					SetForceRegister(true),
					SetRegistrationInfo("https://localhost:8123", "someToken", false)),
				prefs: &agentPreferencesMock{
					AgentRegisteredFunc:     func() bool { return true },
					GetDeviceInfoFunc:       func() *preferences.Device { return &preferences.Device{} },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
				agentUI: &uiMock{
					DisplayRegistrationWindowFunc: func(_ context.Context, _ *preferences.Registration) chan bool {
						doneCh := make(chan bool)
						go func() {
							doneCh <- false
							close(doneCh)
						}()
						return doneCh
					},
				},
			},
		},
		{
			name: "register headless",
			args: args{
				ctx: LoadCtx(ctx,
					SetHeadless(true),
					SetRegistrationInfo("https://localhost:8123", "someToken", false)),
				prefs: &agentPreferencesMock{
					AgentRegisteredFunc:     func() bool { return false },
					GetDeviceInfoFunc:       func() *preferences.Device { return &preferences.Device{} },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
			},
		},
		{
			name: "register",
			args: args{
				ctx: LoadCtx(ctx,
					SetHeadless(false),
					SetRegistrationInfo("https://localhost:8123", "someToken", false)),
				prefs: &agentPreferencesMock{
					AgentRegisteredFunc:     func() bool { return false },
					GetDeviceInfoFunc:       func() *preferences.Device { return &preferences.Device{} },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
				agentUI: &uiMock{
					DisplayRegistrationWindowFunc: func(_ context.Context, _ *preferences.Registration) chan bool {
						doneCh := make(chan bool)
						go func() {
							doneCh <- false
							close(doneCh)
						}()
						return doneCh
					},
				},
			},
		},
		{
			name: "fail",
			args: args{
				ctx: LoadCtx(ctx,
					SetHeadless(true),
					SetRegistrationInfo("https://localhost:8123", "someToken", false)),
				prefs: &agentPreferencesMock{
					AgentRegisteredFunc:     func() bool { return false },
					GetDeviceInfoFunc:       func() *preferences.Device { return &preferences.Device{} },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return errors.New("failed") },
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkRegistration(tt.args.ctx, tt.args.agentUI, tt.args.prefs); (err != nil) != tt.wantErr {
				t.Errorf("Agent.checkRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
