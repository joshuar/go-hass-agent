// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,goconst
package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func newHassPrefs(t *testing.T) *preferences.Hass {
	t.Helper()
	return &preferences.Hass{
		WebhookID: "testWebook",
	}
}

func mockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == hass.RegistrationPath {
			token := request.Header.Get("Authorization")
			if strings.Contains(token, "valid") {
				details := &preferences.Hass{
					WebhookID: "valid",
				}
				body, err := json.Marshal(details)
				assert.NoError(t, err)
				_, err = response.Write(body)
				assert.NoError(t, err)
			} else {
				response.WriteHeader(http.StatusBadRequest)
			}
		}
	}))
}

func TestAgent_saveRegistration(t *testing.T) {
	validPrefs := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "preferences.toml"))

	type fields struct {
		ui            UI
		done          chan struct{}
		prefs         *preferences.Preferences
		logger        *slog.Logger
		id            string
		headless      bool
		forceRegister bool
	}
	type args struct {
		hassPrefs *preferences.Hass
	}
	tests := []struct {
		args    args
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid preferences",
			args:   args{hassPrefs: newHassPrefs(t)},
			fields: fields{prefs: validPrefs},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:            tt.fields.ui,
				done:          tt.fields.done,
				prefs:         tt.fields.prefs,
				logger:        tt.fields.logger,
				id:            tt.fields.id,
				headless:      tt.fields.headless,
				forceRegister: tt.fields.forceRegister,
			}
			if err := agent.saveRegistration(tt.args.hassPrefs); (err != nil) != tt.wantErr {
				t.Errorf("Agent.saveRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.True(t, agent.prefs.Registered)
		})
	}
}

//revive:disable:function-length
func TestAgent_checkRegistration(t *testing.T) {
	server := mockServer(t)

	alreadyRegistered := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "preferences.toml"))
	alreadyRegistered.Registered = true
	alreadyRegistered.Hass.WebhookID = "valid"
	alreadyRegistered.Registration.Server = server.URL
	alreadyRegistered.Registration.Token = "valid"

	headless := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "preferences.toml"))
	headless.Registration.Server = server.URL
	headless.Registration.Token = "valid"

	headlessErr := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "preferences.toml"))
	headlessErr.Registration.Server = server.URL
	headlessErr.Registration.Token = "bad"

	type fields struct {
		ui            UI
		done          chan struct{}
		prefs         *preferences.Preferences
		logger        *slog.Logger
		id            string
		headless      bool
		forceRegister bool
	}
	type args struct {
		trk Tracker
	}
	tests := []struct {
		args    args
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "already registered",
			fields: fields{prefs: alreadyRegistered, id: "go-hass-agent-test"},
		},
		{
			name:   "headless",
			fields: fields{prefs: headless, headless: true, id: "go-hass-agent-test", logger: slog.Default()},
		},
		{
			name:    "headless error",
			fields:  fields{prefs: headlessErr, headless: true, id: "go-hass-agent-test", logger: slog.Default()},
			wantErr: true,
		},
		{
			name:   "force register",
			fields: fields{prefs: alreadyRegistered, headless: true, forceRegister: true, id: "go-hass-agent-test", logger: slog.Default()},
			args:   args{trk: &SensorTrackerMock{ResetFunc: func() {}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:            tt.fields.ui,
				done:          tt.fields.done,
				prefs:         tt.fields.prefs,
				logger:        tt.fields.logger,
				id:            tt.fields.id,
				headless:      tt.fields.headless,
				forceRegister: tt.fields.forceRegister,
			}
			if err := agent.checkRegistration(context.TODO(), tt.args.trk); (err != nil) != tt.wantErr {
				t.Errorf("Agent.checkRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.True(t, agent.prefs.Registered)
				assert.Equal(t, "valid", agent.prefs.Hass.WebhookID)
			}
		})
	}
}

func Test_generateAPIURL(t *testing.T) {
	type args struct {
		prefs  *preferences.Hass
		server string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid cloudhookurl",
			args: args{
				server: "http://localhost",
				prefs: &preferences.Hass{
					CloudhookURL: "http://localhost/cloudhook",
				},
			},
			want: "http://localhost/cloudhook",
		},
		{
			name: "valid remoteuiurl",
			args: args{
				server: "http://localhost",
				prefs: &preferences.Hass{
					RemoteUIURL: "http://localhost/remoteuiurl",
					WebhookID:   "foobar",
				},
			},
			want: "http://localhost/remoteuiurl" + hass.WebHookPath + "foobar",
		},
		{
			name: "webhookid only",
			args: args{
				server: "http://localhost",
				prefs: &preferences.Hass{
					WebhookID: "foobar",
				},
			},
			want: "http://localhost" + hass.WebHookPath + "foobar",
		},
		{
			name: "all defined cloudhookurl",
			args: args{
				server: "http://localhost",
				prefs: &preferences.Hass{
					CloudhookURL: "http://localhost/cloudhook",
					RemoteUIURL:  "http://localhost/remoteuiurl",
					WebhookID:    "foobar",
				},
			},
			want: "http://localhost/cloudhook",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateAPIURL(tt.args.server, tt.args.prefs)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAPIURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateWebsocketURL(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ws conversion",
			args: args{
				host: "http://localhost",
			},
			want: "ws://localhost" + hass.WebsocketPath,
		},
		{
			name: "wss conversion",
			args: args{
				host: "https://localhost",
			},
			want: "wss://localhost" + hass.WebsocketPath,
		},
		{
			name: "ws",
			args: args{
				host: "ws://localhost",
			},
			want: "ws://localhost" + hass.WebsocketPath,
		},
		{
			name: "wss",
			args: args{
				host: "wss://localhost",
			},
			want: "wss://localhost" + hass.WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateWebsocketURL(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateWebsocketURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateWebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
