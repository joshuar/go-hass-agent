// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferences_SetHassPreferences(t *testing.T) {
	appID = "go-hass-agent-test"

	type args struct {
		prefs   *Hass
		options *Registration
		path    string
	}
	tests := []struct {
		args             args
		name             string
		wantAPIURL       string
		wantWebSocketURL string
		wantErr          bool
	}{
		{
			name: "use cloudhook",
			args: args{
				prefs:   &Hass{CloudhookURL: "http://localhost:8123/cloudhook", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost:8123/cloudhook",
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "use remoteui",
			args: args{
				prefs:   &Hass{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost:8123/remoteui/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "ignoreURLs",
			args: args{
				prefs:   &Hass{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret, IgnoreHassURLs: true},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "no cloudhook or remoteui",
			args: args{
				prefs:   &Hass{WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "existing preferences",
			args: args{
				prefs:   &Hass{WebhookID: "newWebHookID"},
				options: &Registration{Server: "http://localhost", Token: "newToken"},
				path:    "testing/data",
			},
			wantAPIURL:       "http://localhost/api/webhook/newWebHookID",
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, checkPath(filepath.Join(tt.args.path, appID)))
			xdg.ConfigHome = tt.args.path
			if err := Load(); err != nil {
				t.Errorf("Preferences.Load() error %v", err)
			}
			if err := SetHassPreferences(tt.args.prefs, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.SetHassPreferences() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantAPIURL, RestAPIURL())
			assert.Equal(t, tt.wantWebSocketURL, WebsocketURL())
		})
	}
}
