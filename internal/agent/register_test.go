// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

//nolint:errcheck
func registrationSuccess(w http.ResponseWriter, _ *http.Request) {
	resp, _ := json.Marshal(&preferences.Hass{})
	w.Write(resp)
}

//nolint:errcheck
func registrationFail(w http.ResponseWriter, _ *http.Request) {
	resp, _ := json.Marshal(&api.ResponseError{
		Code:    http.StatusBadRequest,
		Message: "Bad Request",
	})
	w.WriteHeader(http.StatusBadRequest)
	w.Write(resp)
}

//nolint:errcheck
func alreadyRegistered(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(""))
}

func TestAgent_checkRegistration(t *testing.T) {
	mockUI := &uiMock{
		DisplayRegistrationWindowFunc: func(_ context.Context, _ *preferences.Registration) chan bool {
			doneCh := make(chan bool)
			go func() {
				doneCh <- false
				close(doneCh)
			}()
			return doneCh
		},
	}

	mockPrefs := &registrationPrefsMock{
		AgentRegisteredFunc:     func() bool { return false },
		SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
	}

	type args struct {
		agentUI       ui
		prefs         registrationPrefs
		device        *preferences.Device
		handler       func(http.ResponseWriter, *http.Request)
		forceRegister bool
		headless      bool
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "already registered",
			args: args{
				prefs: &registrationPrefsMock{
					AgentRegisteredFunc:     func() bool { return true },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
				device:  &preferences.Device{},
				handler: alreadyRegistered,
			},
		},
		{
			name: "force registration",
			args: args{
				forceRegister: true,
				prefs: &registrationPrefsMock{
					AgentRegisteredFunc:     func() bool { return true },
					SaveHassPreferencesFunc: func(_ *preferences.Hass, _ *preferences.Registration) error { return nil },
				},
				device:  &preferences.Device{},
				agentUI: mockUI,
				handler: registrationSuccess,
			},
		},
		{
			name: "register headless",
			args: args{
				headless: true,
				prefs:    mockPrefs,
				device:   &preferences.Device{},
				handler:  registrationSuccess,
			},
		},
		{
			name: "register",
			args: args{
				prefs:   mockPrefs,
				device:  &preferences.Device{},
				agentUI: mockUI,
				handler: registrationSuccess,
			},
		},
		{
			name: "fail",
			args: args{
				prefs:   mockPrefs,
				device:  &preferences.Device{},
				agentUI: mockUI,
				handler: registrationFail,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(tt.args.handler))
			defer svr.Close()

			ctx := preferences.AppIDToContext(context.TODO(), "go-hass-agent-test")
			ctx = LoadCtx(ctx,
				SetHeadless(tt.args.headless),
				SetForceRegister(tt.args.forceRegister),
				SetRegistrationInfo(svr.URL, "someToken", false))

			if err := checkRegistration(ctx, tt.args.agentUI, tt.args.device, tt.args.prefs); (err != nil) != tt.wantErr {
				t.Errorf("Agent.checkRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
